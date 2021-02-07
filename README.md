Initial idea

json file => terraform plan
[]resources
resource => go type

Need to be able to specific which values we are interested in.


```
func TestResourceTerraform(t \*t.Testing) {
  want := tfResource{
    type: "nomad_job",
    name: "unit-test",
    count: 2,
  }

  plan, err := planTerraform("dir")
  if err != nil {
    t.Error(error)
  }
  if !plan.Equal(want, got) {
        t.Error(plan.Diff(want,got))
    }

}
```
