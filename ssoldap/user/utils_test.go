package user

import (
	"testing"
)

func TestEIdToIdToEId(t *testing.T) {
	eid := "201507030001"
	id := getIdByEId(eid)
	eid2 := getEIdById(id)
	t.Log(id)
	t.Log(eid2)
	if eid != eid2 {
		t.Error("fail")
	}
}
