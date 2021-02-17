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

	tfPlan := &terraformtest.TFPlan{
		MaxDepth:      1000,
		ItemsMetadata: make(map[string]map[string]gjson.Result),
	}
	want := make(map[string]map[string]gjson.Result)
	want["module.my_module"] = make(map[string]gjson.Result)
	want["abc"] = make(map[string]gjson.Result)
	want["abc"]["name"] = gjson.Result{
		Type:  gjson.String,
		Raw:   `"bogus"`,
		Str:   "bogus",
		Num:   0,
		Index: 48,
	}

	data := []byte(`{
		"planned_values": {
		  "root_module": {
			"child_modules": [
			  {
				"resources": [
				  {
					"name": "bogus",
					"address": "abc"
				  }
				],
				"address": "module.my_module"
			  }
			]
		  }
		}
	  }
	  `)
	tfPlan.Data = data
	tfPlan.Coalesce()
	got := tfPlan.ItemsMetadata
	if !cmp.Equal(want, got) {
		t.Errorf(cmp.Diff(want, got))
	}

}

func TestEqual(t *testing.T) {
	t.Parallel()

	want := terraformtest.TFTestResource{
		Address: "module.nomad_job.nomad_job.test_job",
		Metadata: map[string]string{
			"type": "nomad_job",
			"name": "test_job",
			// "values.name":          "unit-test",
			// "values.datacenters.0": "dc1",
		},
	}
	got, err := terraformtest.NewTerraformTest("testdata/terraform.tfplan")
	if err != nil {
		t.Fatalf("cannot run NewTerraformTest function: %v", err)
	}
	got.Coalesce()

	tfDiff, equal := terraformtest.Equal(want, *got)
	if !equal {
		t.Error(terraformtest.Diff(tfDiff))
	}
}

func TestTFAWS101NatEIPOne(t *testing.T) {
	t.Parallel()

	want := terraformtest.TFTestResource{
		Address: "module.vpc.aws_eip.nat[0]",
		Metadata: map[string]string{
			"type": "aws_eip",
			"name": "nat",
			// "values.vpc":            "true",
			// "values.tags.Terraform": "true",
			// "values.timeouts":       "",
		},
	}

	got, err := terraformtest.NewTerraformTest("testdata/terraform-aws-101.tfplan.json")
	if err != nil {
		t.Fatalf("cannot read terraform plan: %v", err)
	}
	got.Coalesce()

	tfDiff, equal := terraformtest.Equal(want, *got)
	if !equal {
		t.Error(terraformtest.Diff(tfDiff))
	}
}

func TestTFAWS101DBOptionGroup(t *testing.T) {
	t.Parallel()

	want := terraformtest.TFTestResource{
		Address: "module.db.module.db_option_group.aws_db_option_group.this[0]",
		Metadata: map[string]string{
			"type": "aws_db_option_group",
			"name": "this",
			// "values.engine_name": "mysql",
			// "values.major_engine_version": "5.7",
			// "values.option.0.option_name": "MARIADB_AUDIT_PLUGIN",
		},
	}

	got, err := terraformtest.NewTerraformTest("testdata/terraform-aws-101.tfplan.json")
	if err != nil {
		t.Fatalf("cannot read terraform plan: %v", err)
	}
	got.Coalesce()

	tfDiff, equal := terraformtest.Equal(want, *got)
	if !equal {
		t.Error(terraformtest.Diff(tfDiff))
	}
}
