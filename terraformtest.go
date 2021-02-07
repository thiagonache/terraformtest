package terraformtest

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/google/go-cmp/cmp"
	"github.com/savaki/jq"
)

func CountNumberResources(tfPlan []byte, jqPath string) (int, error) {
	tfPlan, err := ExtractPlanData(tfPlan, jqPath)
	if err != nil {
		return 0, fmt.Errorf("cannot extract data from plan using %q query path", jqPath)
	}
	fmt.Println(cmp.Diff(string(tfPlan), "datacenters"))

	return 0, nil
}
func ExtractPlanData(tfPlan []byte, jqPath string) ([]byte, error) {
	jqOp, err := jq.Parse(jqPath)
	if err != nil {
		return nil, fmt.Errorf("invalid jq query: %q", jqPath)
	}

	value, err := jqOp.Apply(tfPlan)
	if err != nil {
		return nil, fmt.Errorf("cannot apply query %q on tfPlan data", jqPath)
	}
	return value, nil
}

func ParseJSON(path string) ([]byte, error) {
	f, err := os.Open(path)
	if err != nil {
		log.Fatalf("cannot open file: %s", path)
	}
	reader := bufio.NewReader(f)
	jsonBytes, err := ioutil.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("cannot read data from IO Reader: %v", err)
	}
	return jsonBytes, nil

}
