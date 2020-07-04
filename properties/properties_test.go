package properties

import (
	"testing"
)

func TestAddingRemovingProperties(t *testing.T) {
	p := NewPropertyMap()

	dontPanic := p.GetProperty("non")
	if dontPanic != nil {
		t.Fatal("Something is wrong, should be nil")
	}

	err := p.SetProperty("someConfig", "123123")
	if err != ErrNoSuchProperty {
		t.Fatal("Should have failed to set property value if its not added before hand")
	}

	p.AddProperty("someConfig", "test", false)
	something := p.GetProperty("someConfig")

	if something == nil {
		t.Fatal("Should have found a Property")
	}

	p.RemoveProperty("someConfig")

	s2 := p.GetProperty("someConfig")
	if s2 != nil {
		t.Fatal("Shouldnt have found any config after removal")
	}
}
func TestValidation(t *testing.T) {
	p := NewPropertyMap()

	p.AddProperty("integer", "an integer valued prop", true)
	err := p.SetProperty("integer", 10)
	if err != nil {
		t.Fatal("integer should have been able to set value", err)
	}
	p.AddProperty("string", "an string valued property", false)
	p.SetProperty("string", "HelloWorld")

	valid, _ := p.ValidateProperties()
	if !valid {
		t.Fatalf("Should have all the needed Properties")
	}

	p.SetProperty("integer", nil)

	valid, missing := p.ValidateProperties()
	if valid {
		t.Fatalf("Should have been invalid, this is the result: %v", missing)
	}
	if len(missing) != 1 {
		t.Fatalf("should have 1 value returned as missing property: %v", missing)
	}
}
func TestReflection(t *testing.T) {
	p := NewPropertyMap()

	p.AddProperty("integer", "an int property", false)
	p.AddProperty("string", "a string prop", false)
	p.SetProperty("integer", 10)
	p.SetProperty("string", "HelloWorld")

	intProp := p.GetProperty("integer")
	strProp := p.GetProperty("string")
	if intProp == nil || strProp == nil {
		t.Fatal("failed to extract needed properties for reflection testing")
	}
	_, _ = intProp.Int()
	_ = strProp.String()

	intAsString := intProp.String()

	t.Log(intAsString)

}
