package fetch

import "testing"

var endp GdcEndpoints

func TestGdc(t *testing.T) {
	//endp.Projects = "1"
	//endp.Status = "1"
	endp.Cases = true
	endp.Files = true
	endp.Annotations = true
	Gdc(endp, "", 2, 2, 2, false)
}
