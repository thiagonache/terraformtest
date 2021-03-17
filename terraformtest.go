package terraformtest

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/tidwall/gjson"
)

type loopControl struct {
	curDepth, maxDepth          int
	curItemIndex, curItemSubKey string
	prevItemIndex               string
}

// Plan is the main struct containing the Plan data
type Plan struct {
	Data        []byte
	loopControl loopControl
	Resources   ResourceSet
}

type compDiff struct {
	items []compDiffItem
}

type compDiffItem struct {
	got, key, want string
}

// Resource represents a resource being tested
type Resource struct {
	Address  string
	Metadata map[string]string
	Values   map[string]string
}

// ResourceSet stores the resources (items) and diff of the plan file.
type ResourceSet struct {
	Resources map[string]map[string]map[string]string
	CompDiff  compDiff
}

// ReadPlan takes the plan's file path in JSON format and returns a pointer to a
// Plan object and an error.
func ReadPlan(planPath string) (*Plan, error) {
	tf := &Plan{
		loopControl: loopControl{maxDepth: 10},
		Resources: ResourceSet{
			Resources: map[string]map[string]map[string]string{},
			CompDiff:  compDiff{},
		},
	}

	f, err := os.Open(planPath)
	if err != nil {
		return tf, fmt.Errorf("cannot open file %s: %v", planPath, err)
	}
	defer f.Close()

	plan, err := io.ReadAll(f)
	if err != nil {
		return tf, fmt.Errorf("cannot read data from IO Reader: %v", err)
	}

	tf.Data = plan
	tf.ResourceSet()

	return tf, nil
}

// Diff iterates over CompDiff items to concatenate all errors by new line.
func (rs ResourceSet) Diff() string {
	var stringDiff string
	for _, diff := range rs.CompDiff.items {
		stringDiff += fmt.Sprintf(`key %q: want %q, got %q\n`, diff.key, diff.want, diff.got)
	}
	return stringDiff
}

// ResourceSet transform the multi level json into one big object to make
// queries easier.
func (tfPlan *Plan) ResourceSet() {
	rootModule := gjson.GetBytes(tfPlan.Data, `planned_values.root_module|@pretty:{"sortKeys":true}`)
	rootModule.ForEach(tfPlan.transform)
}

func (tfPlan *Plan) transform(key, value gjson.Result) bool {
	if tfPlan.loopControl.curDepth > tfPlan.loopControl.maxDepth {
		fmt.Println("MaxDepth reached")
		return false
	}

	switch key.String() {
	case "resources":
		tfPlan.loopControl.prevItemIndex = "resources"
		tfPlan.loopControl.curDepth++
		for _, child := range value.Array() {
			child.ForEach(tfPlan.transform)
		}
	case "child_modules":
		tfPlan.loopControl.prevItemIndex = "child_modules"
		tfPlan.loopControl.curDepth++
		for _, child := range value.Array() {
			child.ForEach(tfPlan.transform)
		}
	case "values":
		tfPlan.loopControl.curItemSubKey = "Values"
		_, ok := tfPlan.Resources.Resources[tfPlan.loopControl.curItemIndex]
		if !ok {
			tfPlan.Resources.Resources[tfPlan.loopControl.curItemIndex] = map[string]map[string]string{}
		}
		tfPlan.Resources.Resources[tfPlan.loopControl.curItemIndex][tfPlan.loopControl.curItemSubKey] = map[string]string{}
		value.ForEach(tfPlan.transform)
	case "address":
		// We are only interested in addresses of resources
		if tfPlan.loopControl.prevItemIndex != "resources" {
			break
		}
		tfPlan.loopControl.curItemSubKey = "Metadata"
		tfPlan.loopControl.curItemIndex = value.String()
		_, ok := tfPlan.Resources.Resources[tfPlan.loopControl.curItemIndex]
		if !ok {
			tfPlan.Resources.Resources[tfPlan.loopControl.curItemIndex] = map[string]map[string]string{}
		}
		tfPlan.Resources.Resources[tfPlan.loopControl.curItemIndex][tfPlan.loopControl.curItemSubKey] = map[string]string{}

	default:
		value := normalizeItem(value.String())
		tfPlan.Resources.Resources[tfPlan.loopControl.curItemIndex][tfPlan.loopControl.curItemSubKey][key.String()] = value
	}

	return true
}

func normalizeItem(item string) string {
	item = strings.ReplaceAll(item, "\n", "")
	item = strings.ReplaceAll(item, " ", "")
	return item
}

func (rs *ResourceSet) newCompDiffItem(key, want, got string) {
	item := compDiffItem{
		got:  got,
		key:  key,
		want: want,
	}
	rs.CompDiff.items = append(rs.CompDiff.items, item)
}

// Contains check if a resource exist in the resourceSet.
func (rs *ResourceSet) Contains(r Resource) bool {
	metadata, ok := rs.Resources[r.Address]["Metadata"]
	if !ok {
		rs.newCompDiffItem(r.Address, "exist", "nil")
		return false
	}
	for k, v := range r.Metadata {
		valueFound, ok := metadata[k]
		if !ok {
			rs.newCompDiffItem(k, "exist", "nil")
			return false
		}
		v = normalizeItem(v)
		if valueFound != v {
			rs.newCompDiffItem(k, v, valueFound)
			return false
		}
	}

	values, ok := rs.Resources[r.Address]["Values"]
	if !ok {
		rs.newCompDiffItem(r.Address, "exist", "nil")
		return false
	}
	for k, v := range r.Values {
		valueFound, ok := values[k]
		if !ok {
			rs.newCompDiffItem(k, "exist", "nil")
			return false
		}
		v = normalizeItem(v)
		if valueFound != v {
			rs.newCompDiffItem(k, v, valueFound)
			return false
		}
	}
	return true
}

// Equal check if all resources exist in the resourceSet and vice-versa.
func Equal(resources []Resource, rs *ResourceSet) bool {
	resourcesRS := map[string]struct{}{}
	for _, r := range resources {
		resourcesRS[r.Address] = struct{}{}
		rsItem, ok := rs.Resources[r.Address]
		if !ok {
			rs.newCompDiffItem(r.Address, "exist in plan", "nil")
			return false
		}

		for k, v := range r.Metadata {
			valueFound, ok := rsItem["Metadata"][k]
			if !ok {
				rs.newCompDiffItem(r.Address, "exist in plan", "nil")
				return false
			}
			v = normalizeItem(v)
			if valueFound != v {
				rs.newCompDiffItem(k, v, valueFound)
				return false
			}
		}

		for k, v := range r.Values {
			valueFound, ok := rsItem["Values"][k]
			if !ok {
				rs.newCompDiffItem(r.Address, "exist in plan", "nil")
				return false
			}
			v = normalizeItem(v)
			if valueFound != v {
				rs.newCompDiffItem(k, v, valueFound)
				return false
			}
		}
	}

	for k := range rs.Resources {
		_, ok := resourcesRS[k]
		if !ok {
			rs.newCompDiffItem(k, "exist in resources", "nil")
			return false
		}
	}
	return true
}
