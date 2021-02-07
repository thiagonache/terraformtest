json file => terraform plan
[]resources
resource => go type

be able to specific which values we are interested in.

func TestResourceTerraform(t \*t.Testing) {
want := tfResource{
type: "nomad_job",
name: "unit-test",
count: 2,
}

    plan, err := planTerraform("dir")
    if err != nil {}
    if !plan.Equal(want, got) {
        fmt.Println(plan.Diff(want,got))
    }

}
