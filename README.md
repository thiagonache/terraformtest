# Unit testing terraform (WIP)

![Go](https://github.com/thiagonache/terraformtest/workflows/Go/badge.svg?branch=master)
[![Go Report Card](https://goreportcard.com/badge/github.com/thiagonache/terraformtest)](https://goreportcard.com/report/github.com/thiagonache/terraformtest)

## Testing your plan

1. Read the json file (mandatory)

   ```go
   p, err := terraformtest.ReadPlan("path/to/terraform.plan.json")
       if err != nil {
       t.Fatal(err)
   }
   ```

1. Test number of resources in the plan file

   ```go
   items := p.PlanResourcesSet.Items
   if len(items) < wantNumResources {
       t.Errorf("want %d resources in plan, got %d", wantNumResources,  len(items))
   }
   ```

1. Test if plan contains a resource

   ```go
   wantRes := terraformtest.TestResource{
       Address: "module.nomad_job.nomad_job.test_job",
       Metadata: map[string]string{
           "type": "nomad_job",
           "name": "test_job",
       },
       Values: map[string]string{
           "name": "unit-test",
       },
   }

   if !gotRS.Contains(wantRes) {
       t.Error(gotRS.Diff())
   }
   ```

1. Test if plan is equal to a group of resources (WIP)
