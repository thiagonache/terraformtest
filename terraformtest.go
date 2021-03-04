package terraformtest

import (
	"fmt"
	"io"
	"os"

	"github.com/tidwall/gjson"
)

// loopControl is a struct containing items to control for loop to process json file
type loopControl struct {
	curDepth, maxDepth          int
	curItemIndex, curItemSubKey string
	prevItemIndex               string
}

// Plan is the main struct containing the Plan data
type Plan struct {
	Data    []byte
	loopControl loopControl
	Resources   ResourceSet
}

// compDiff is a struct containing slice of CompDiffItem
type compDiff struct {
	items []compDiffItem
}

// compDiffItem is a struct containing got, key and want values for the diff
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
	Resources map[string]map[string]map[string]gjson.Result
	CompDiff  compDiff
}

// ReadPlan takes the plan's file path in JSON format and returns a pointer to a
// Test object and an error.
func ReadPlan(planPath string) (*Plan, error) {
	tf := &Plan{
		loopControl: loopControl{maxDepth: 10},
		Resources: ResourceSet{
			Resources: map[string]map[string]map[string]gjson.Result{},
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
	rootModule.ForEach(tfPlan.Transform)
}

// Transform iterates over TF Plan in JSON format to produce a single level map
// of resources.
func (tfPlan *Plan) Transform(key, value gjson.Result) bool {
	if tfPlan.loopControl.curDepth > tfPlan.loopControl.maxDepth {
		fmt.Println("MaxDepth reached")
		return false
	}

	switch key.String() {
	case "resources":
		tfPlan.loopControl.prevItemIndex = "resources"
		tfPlan.loopControl.curDepth++
		for _, child := range value.Array() {
			child.ForEach(tfPlan.Transform)
		}
	case "child_modules":
		tfPlan.loopControl.prevItemIndex = "child_modules"
		tfPlan.loopControl.curDepth++
		for _, child := range value.Array() {
			child.ForEach(tfPlan.Transform)
		}
	case "values":
		tfPlan.loopControl.curItemSubKey = "Values"
		_, ok := tfPlan.Resources.Resources[tfPlan.loopControl.curItemIndex]
		if !ok {
			tfPlan.Resources.Resources[tfPlan.loopControl.curItemIndex] = map[string]map[string]gjson.Result{}
		}
		tfPlan.Resources.Resources[tfPlan.loopControl.curItemIndex][tfPlan.loopControl.curItemSubKey] = map[string]gjson.Result{}
		value.ForEach(tfPlan.Transform)
	case "address":
		// We are only interested in addresses of resources
		if tfPlan.loopControl.prevItemIndex != "resources" {
			break
		}
		tfPlan.loopControl.curItemSubKey = "Metadata"
		tfPlan.loopControl.curItemIndex = value.String()
		_, ok := tfPlan.Resources.Resources[tfPlan.loopControl.curItemIndex]
		if !ok {
			tfPlan.Resources.Resources[tfPlan.loopControl.curItemIndex] = map[string]map[string]gjson.Result{}
		}
		tfPlan.Resources.Resources[tfPlan.loopControl.curItemIndex][tfPlan.loopControl.curItemSubKey] = map[string]gjson.Result{}

	default:
		tfPlan.Resources.Resources[tfPlan.loopControl.curItemIndex][tfPlan.loopControl.curItemSubKey][key.String()] = value
	}

	return true
}

// NewCompDiffItem abstracts to append a new item to the slice of diffs.
func (rs *ResourceSet) NewCompDiffItem(key, want, got string) {
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
		rs.NewCompDiffItem(r.Address, "exist", "nil")
		return false
	}
	for k, v := range r.Metadata {
		valueFound, ok := metadata[k]
		if !ok {
			rs.NewCompDiffItem(k, "exist", "nil")
			return false
		}
		if valueFound.String() != v {
			rs.NewCompDiffItem(k, v, valueFound.String())
			return false
		}
	}

	values, ok := rs.Resources[r.Address]["Values"]
	if !ok {
		rs.NewCompDiffItem(r.Address, "exist", "nil")
		return false
	}
	for k, v := range r.Values {
		valueFound, ok := values[k]
		if !ok {
			rs.NewCompDiffItem(k, "exist", "nil")
			return false
		}
		if valueFound.String() != v {
			rs.NewCompDiffItem(k, v, valueFound.String())
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
			rs.NewCompDiffItem(r.Address, "exist in plan", "nil")
			return false
		}

		for k, v := range r.Metadata {
			valueFound, ok := rsItem["Metadata"][k]
			if !ok {
				rs.NewCompDiffItem(r.Address, "exist in plan", "nil")
				return false
			}
			if valueFound.String() != v {
				rs.NewCompDiffItem(k, v, valueFound.String())
				return false
			}
		}

		for k, v := range r.Values {
			valueFound, ok := rsItem["Values"][k]
			if !ok {
				rs.NewCompDiffItem(r.Address, "exist in plan", "nil")
				return false
			}
			if valueFound.String() != v {
				rs.NewCompDiffItem(k, v, valueFound.String())
				return false
			}
		}
	}

	for k := range rs.Resources {
		_, ok := resourcesRS[k]
		if !ok {
			rs.NewCompDiffItem(k, "exist in resources", "nil")
			return false
		}
	}
	return true
}
