package terraformtest_test

import (
	"terraformtest"
	"testing"
)

func TestReadTfPlan(t *testing.T) {
	t.Parallel()

	wantBiggerThan := 2000
	tfPlan, err := terraformtest.ReadTfPlan("terraform.tfplan")
	if err != nil {
		t.Error(err)
	}
	if wantBiggerThan >= len(tfPlan.Data) {
		t.Errorf("want json minimum size in bytes of %d but got %d", wantBiggerThan, len(tfPlan.Data))
	}
}

func TestEqual(t *testing.T) {
	t.Parallel()

	want := terraformtest.TFResource{
		Filter: `planned_values.root_module.child_modules.#.resources`,
		Check: map[string]string{
			"0.0.address":              "module.nomad_job.nomad_job.test_job",
			"0.0.type":                 "nomad_job",
			"0.0.values.name":          "unit-test",
			"0.0.values.datacenters.0": "dc1",
		},
	}
	got, err := terraformtest.ReadTfPlan("terraform.tfplan")
	if err != nil {
		t.Fatalf("cannot read terraform plan: %v", err)
	}

	tfDiff, equal := terraformtest.Equal(want, got)
	if !equal {
		t.Error(terraformtest.OutputDiff(tfDiff))
	}
}

func TestTFAWS101NatEIPOne(t *testing.T) {
	t.Parallel()

	want := terraformtest.TFResource{
		Filter: `planned_values.root_module.child_modules.#.resources`,
		Check: map[string]string{
			"0.0.address":               "module.vpc.aws_eip.nat[0]",
			"0.0.type":                  "aws_eip",
			"0.0.values.vpc":            "true",
			"0.0.values.tags.Terraform": "true",
			"0.0.values.timeouts":       "",
		},
	}

	got, err := terraformtest.ReadTfPlan("testdata/terraform-aws-101.tfplan.json")
	if err != nil {
		t.Fatalf("cannot read terraform plan: %v", err)
	}

	tfDiff, equal := terraformtest.Equal(want, got)
	if !equal {
		t.Error(terraformtest.OutputDiff(tfDiff))
	}
}

func TestTFAWS101DBOptionGroup(t *testing.T) {
	t.Parallel()

	want := terraformtest.TFResource{
		Filter: `planned_values.root_module.child_modules.#.child_modules.#.resources`,
		Check: map[string]string{
			"0.0.0.address":                     "module.db.module.db_option_group.aws_db_option_group.this[0]",
			"0.0.0.type":                        "aws_db_option_group",
			"0.0.0.values.engine_name":          "mysql",
			"0.0.0.values.major_engine_version": "5.7",
			"0.0.0.values.option.0.option_name": "MARIADB_AUDIT_PLUGIN",
		},
	}

	got, err := terraformtest.ReadTfPlan("testdata/terraform-aws-101.tfplan.json")
	if err != nil {
		t.Fatalf("cannot read terraform plan: %v", err)
	}

	tfDiff, equal := terraformtest.Equal(want, got)
	if !equal {
		t.Error(terraformtest.OutputDiff(tfDiff))
	}
}
