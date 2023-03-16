package alerts

import (
	"fmt"
)

type Errors []error

func (es *Errors) collect(e error) { *es = append(*es, e) }

func (es Errors) Error() (err string) {
	err = "errors:\n"
	for i, e := range es {
		err += fmt.Sprintf("Error %d: %s\n", i, e.Error())
	}
	return err
}
