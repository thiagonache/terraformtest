package terraformtest_test

import (
	"terraformtest"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/tidwall/gjson"
)

func TestReadPlanFile(t *testing.T) {
	t.Parallel()

	wantLen := 9028
	tfPlan, err := terraformtest.New("testdata/terraform.plan.json")
	if err != nil {
		t.Fatal(err)
	}

	if wantLen != len(tfPlan.Data) {
		t.Errorf("want json size in bytes of %d but got %d", wantLen, len(tfPlan.Data))
	}
}

// func TestRefactoredAPI(t *testing.T) {
// 	t.Parallel()
// 	p, err := terraformtest.ReadPlan("testdata/terraform.plan.json")
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	gotRS := p.ResourceSet()
// 	wantNumResources := 3
// 	if len(gotRS) < wantNumResources {
// 		t.Errorf("want %d resources in plan, got %d", wantNumResources, len(gotRS))
// 	}
// 	wantRes := terraformtest.TFTestResource{
// 		Address: "module.nomad_job.nomad_job.test_job",
// 		Metadata: map[string]string{
// 			"type": "nomad_job",
// 			"name": "test_job",
// 			// "values.name":          "unit-test",
// 			// "values.datacenters.0": "dc1",
// 		},
// 	}
// 	if !gotRS.Contains(wantRes) {
// 		t.Errorf("want resource set to contain resource %q, but it didn't:", wantRes.Address)
// 	}
// 	wantRS := []terraformtest.TFTestResource{
// 		{
// 			Address: "module.nomad_job.nomad_job.test_job",
// 			Metadata: map[string]string{
// 				"type": "nomad_job",
// 				"name": "test_job",
// 				// "values.name":          "unit-test",
// 				// "values.datacenters.0": "dc1",
// 			},
// 		},
// 		{
// 			Address: "module.nomad_job.nomad_job.test_job2",
// 			Metadata: map[string]string{
// 				"type": "nomad_job",
// 				"name": "test_job",
// 				// "values.name":          "unit-test",
// 				// "values.datacenters.0": "dc1",
// 			},
// 		},
// 		{
// 			Address: "module.nomad_job.nomad_job.test_job3",
// 			Metadata: map[string]string{
// 				"type": "nomad_job",
// 				"name": "test_job",
// 				// "values.name":          "unit-test",
// 				// "values.datacenters.0": "dc1",
// 			},
// 		},
// 	}
// 	if !terraformtest.Equal(wantRS, gotRS) {
// 		t.Error(terraformtest.Diff(wantRS, gotRS))
// 	}
// }

func TestCoalescePlan(t *testing.T) {
	t.Parallel()

	tfPlan := &terraformtest.TFPlan{
		LoopControl: terraformtest.LoopControl{MaxDepth: 10},
		Items:       map[string]terraformtest.TFResultResource{},
	}
	want := map[string]terraformtest.TFResultResource{}
	want["Metadata"] = map[string]map[string]gjson.Result{}
	want["Metadata"]["abc"] = map[string]gjson.Result{}
	want["Metadata"]["module.my_module"] = map[string]gjson.Result{}
	want["Metadata"]["abc"]["name"] = gjson.Result{
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
	got := tfPlan.Items
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
	got, err := terraformtest.New("testdata/terraform.plan.json")
	if err != nil {
		t.Fatalf("cannot run New function: %v", err)
	}

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

	got, err := terraformtest.New("testdata/terraform-aws-101.plan.json")
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
		Address: "module.db.module.db_option_group.aws_db_option_group.this[0]",
		Metadata: map[string]string{
			"type": "aws_db_option_group",
			"name": "this",
			// "values.engine_name": "mysql",
			// "values.major_engine_version": "5.7",
			// "values.option.0.option_name": "MARIADB_AUDIT_PLUGIN",
		},
	}

	got, err := terraformtest.New("testdata/terraform-aws-101.plan.json")
	if err != nil {
		t.Fatalf("cannot read terraform plan: %v", err)
	}

	tfDiff, equal := terraformtest.Equal(want, *got)
	if !equal {
		t.Error(terraformtest.Diff(tfDiff))
	}
}
