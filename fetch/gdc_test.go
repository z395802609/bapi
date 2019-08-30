package fetch

import (
	"testing"

	"github.com/Miachol/bapi/types"
)

var endp types.GdcEndpoints

func TestGdc(t *testing.T) {
	//endp.Projects = "1"
	//endp.Status = "1"
	endp.Cases = true
	endp.Files = true
	endp.Annotations = true
	var bapiClis = &types.BapiClisT{
		Retries: 5,
	}
	Gdc(&endp, bapiClis)
}
