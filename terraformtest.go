package terraformtest

import (
	"fmt"
	"io"
	"os"

	"github.com/tidwall/gjson"
)

// LoopControl is a struct containing items to control for loop to process json file
type LoopControl struct {
	CurDepth, MaxDepth          int
	CurItemIndex, CurItemSubKey string
	PrevItemIndex               string
}

// Test is the main struct containing the test data
type Test struct {
	PlanData    []byte
	LoopControl LoopControl
	Resources   ResourceSet
}

// CompDiff is a struct containing slice of CompDiffItem
type CompDiff struct {
	Items []CompDiffItem
}

// CompDiffItem is a struct containing got, key and want values for the diff
type CompDiffItem struct {
	Got, Key, Want string
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
	CompDiff  CompDiff
}

func ReadPlan(planPath string) (*Test, error) {
	tf := &Test{
		LoopControl: LoopControl{MaxDepth: 10},
		Resources: ResourceSet{
			Resources: map[string]map[string]map[string]gjson.Result{},
			CompDiff:  CompDiff{},
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

	tf.PlanData = plan
	tf.ResourceSet()

	return tf, nil
}

func (rs ResourceSet) Diff() string {
	var stringDiff string
	for _, diff := range rs.CompDiff.Items {
		stringDiff += fmt.Sprintf(`key %q: want %q, got %q\n`, diff.Key, diff.Want, diff.Got)
	}
	return stringDiff
}

// ResourceSet transform the multi level json into one big object to make queries easier
func (tfPlan *Test) ResourceSet() {
	rootModule := gjson.GetBytes(tfPlan.PlanData, `planned_values.root_module|@pretty:{"sortKeys":true}`)
	rootModule.ForEach(tfPlan.transform)
}

func (tfPlan *Test) transform(key, value gjson.Result) bool {
	if tfPlan.LoopControl.CurDepth > tfPlan.LoopControl.MaxDepth {
		fmt.Println("MaxDepth reached")
		return false
	}

	switch key.String() {
	case "resources":
		tfPlan.LoopControl.PrevItemIndex = "resources"
		tfPlan.LoopControl.CurDepth++
		for _, child := range value.Array() {
			child.ForEach(tfPlan.transform)
		}
	case "child_modules":
		tfPlan.LoopControl.PrevItemIndex = "child_modules"
		tfPlan.LoopControl.CurDepth++
		for _, child := range value.Array() {
			child.ForEach(tfPlan.transform)
		}
	case "values":
		tfPlan.LoopControl.CurItemSubKey = "Values"
		_, ok := tfPlan.Resources.Resources[tfPlan.LoopControl.CurItemIndex]
		if !ok {
			tfPlan.Resources.Resources[tfPlan.LoopControl.CurItemIndex] = map[string]map[string]gjson.Result{}
		}
		tfPlan.Resources.Resources[tfPlan.LoopControl.CurItemIndex][tfPlan.LoopControl.CurItemSubKey] = map[string]gjson.Result{}
		value.ForEach(tfPlan.transform)
	case "address":
		// We are only interested in addresses of resources
		if tfPlan.LoopControl.PrevItemIndex != "resources" {
			break
		}
		tfPlan.LoopControl.CurItemSubKey = "Metadata"
		tfPlan.LoopControl.CurItemIndex = value.String()
		_, ok := tfPlan.Resources.Resources[tfPlan.LoopControl.CurItemIndex]
		if !ok {
			tfPlan.Resources.Resources[tfPlan.LoopControl.CurItemIndex] = map[string]map[string]gjson.Result{}
		}
		tfPlan.Resources.Resources[tfPlan.LoopControl.CurItemIndex][tfPlan.LoopControl.CurItemSubKey] = map[string]gjson.Result{}

	default:
		tfPlan.Resources.Resources[tfPlan.LoopControl.CurItemIndex][tfPlan.LoopControl.CurItemSubKey][key.String()] = value
		//fmt.Printf("Add key %v and value %v into %v into %v\n\n", key, value, tfPlan.LoopControl.CurItemIndex, tfPlan.LoopControl.CurItemSubKey)
	}

	return true
}

func (rs *ResourceSet) NewCompDiffItem(key, want, got string) {
	item := CompDiffItem{
		Got:  got,
		Key:  key,
		Want: want,
	}
	rs.CompDiff.Items = append(rs.CompDiff.Items, item)
}

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
