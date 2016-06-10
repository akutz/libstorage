package types

import "testing"

func TestParentConfigKeys(t *testing.T) {
	for k := range ParentConfigKeys() {
		t.Logf("%[1]d, %[1]s", k)
	}
}

func TestValueConfigKeys(t *testing.T) {
	for k := range ValueConfigKeys() {
		t.Logf("%s", k)
	}
}

func TestGetConfigSectionInfo(t *testing.T) {
	vk, sk, ok := GetConfigSectionInfo(ConfigRoot)
	if !ok {
		t.FailNow()
	}
	t.Logf("valKeys=%v, subKeys=%v", vk, sk)

	vk, sk, ok = GetConfigSectionInfo(ConfigIGVolOpsCreateDefault)
	if !ok {
		t.FailNow()
	}
	t.Logf("valKeys=%v, subKeys=%v", vk, sk)
}
