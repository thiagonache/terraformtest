package terraformtest_test

import (
	"terraformtest"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/tidwall/gjson"
)

func TestReadPlanFile(t *testing.T) {
	t.Parallel()

	want := 9028
	tfPlan, err := terraformtest.NewTerraformTest("testdata/terraform.tfplan")
	if err != nil {
		t.Fatal(err)
	}

	if want != len(tfPlan.Data) {
		t.Errorf("want json size in bytes of %d but got %d", want, len(tfPlan.Data))
	}
}

func TestCoalescePlan(t *testing.T) {
	t.Parallel()
	tfPlan := terraformtest.TFPlan{
		MaxDepth: 10,
	}
	want := make(map[string]map[string]gjson.Result)
	want["abc"] = make(map[string]gjson.Result)
	want["abc"]["name"] = gjson.Result{
		Type:  gjson.String,
		Raw:   `"bogus"`,
		Str:   "bogus",
		Num:   0,
		Index: 0,
	}

	data := []byte(`{
		"planned_values": {
		  "root_module": {
			"child_modules": [
			  {
				"resources": [
				  {
					"address": "abc",
					"name": "bogus"
				  }
				],
				"address": "module.my_module"
			  }
			]
		  }
		}
	  }
	  `)
	rootModule := gjson.GetBytes(data, `planned_values.root_module`)
	for k, v := range rootModule.Map() {
		terraformtest.CoalescePlan(&tfPlan, k, v)
	}
	got := tfPlan.Items
	if !cmp.Equal(want, got) {
		t.Errorf(cmp.Diff(want, got))
	}

}
func TestEqual(t *testing.T) {
	t.Parallel()

	want := terraformtest.TFTestResource{
		Filter: `planned_values.root_module.child_modules.#.resources`,
		Check: map[string]string{
			"0.0.address":              "module.nomad_job.nomad_job.test_job",
			"0.0.type":                 "nomad_job",
			"0.0.values.name":          "unit-test",
			"0.0.values.datacenters.0": "dc1",
		},
	}
	got, err := terraformtest.NewTerraformTest("testdata/terraform.tfplan")
	if err != nil {
		t.Fatalf("cannot read terraform plan: %v", err)
	}

	tfDiff, equal := terraformtest.Equal(want, *got)
	if !equal {
		t.Error(terraformtest.Diff(tfDiff))
	}
}

func TestTFAWS101NatEIPOne(t *testing.T) {
	t.Parallel()

	want := terraformtest.TFTestResource{
		Filter: `planned_values.root_module.child_modules.#.resources`,
		Check: map[string]string{
			"0.0.address":               "module.vpc.aws_eip.nat[0]",
			"0.0.type":                  "aws_eip",
			"0.0.values.vpc":            "true",
			"0.0.values.tags.Terraform": "true",
			"0.0.values.timeouts":       "",
		},
	}

	got, err := terraformtest.NewTerraformTest("testdata/terraform-aws-101.tfplan.json")
	if err != nil {
		t.Fatalf("cannot read terraform plan: %v", err)
	}

	tfDiff, equal := terraformtest.Equal(want, *got)
	if !equal {
		t.Error(terraformtest.Diff(tfDiff))
	}
}

func TestTFAWS101DBOptionGroup(t *testing.T) {
	t.Parallel()

	want := terraformtest.TFTestResource{
		Filter: `planned_values.root_module.child_modules.#.child_modules.#.resources`,
		Check: map[string]string{
			"0.0.0.address":                     "module.db.module.db_option_group.aws_db_option_group.this[0]",
			"0.0.0.type":                        "aws_db_option_group",
			"0.0.0.values.engine_name":          "mysql",
			"0.0.0.values.major_engine_version": "5.7",
			"0.0.0.values.option.0.option_name": "MARIADB_AUDIT_PLUGIN",
		},
	}

	got, err := terraformtest.NewTerraformTest("testdata/terraform-aws-101.tfplan.json")
	if err != nil {
		t.Fatalf("cannot read terraform plan: %v", err)
	}

	tfDiff, equal := terraformtest.Equal(want, *got)
	if !equal {
		t.Error(terraformtest.Diff(tfDiff))
	}
}
