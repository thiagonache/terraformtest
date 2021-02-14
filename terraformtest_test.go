package terraformtest_test

import (
	"terraformtest"
	"testing"
	"testdata/terraform.tfplan"
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
		Metadata: map[string]string{
			"name": "test_job",
			"type": "nomad_job",
		},
		Total: 1,
		Values: map[string]string{
			"name":          "unit-test",
			"datacenters.0": "dc1",
		},
	}
	got, err := terraformtest.ReadTfPlan("terraform.tfplan")
	if err != nil {
		t.Fatalf("cannot read terraform plan: %v", err)
	}

	if !terraformtest.Equal(want, got) {
		t.Error("not equal")
	}
}
