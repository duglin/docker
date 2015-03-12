package config

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"testing"
)

// Test config structs
type PersonType struct {
	Name string
	Age  int
}

type GroupType struct {
	JsonConfig
	Name      string
	AnInt     int
	ABool     bool
	Manager   PersonType
	StrArray  []string
	People    []PersonType
	PeopleMap map[string]PersonType
}

func NewGroupType() *GroupType {
	g := &GroupType{}
	g.MakeConfig(g)
	return g
}

func TestNoConfig(t *testing.T) {
	fmt.Sprintf("hi")

	ctx, err := NewContext(nil)
	if err != nil {
		t.Fatalf("Error creating context: %q", err)
	}
	defer ctx.Delete()

	data := NewGroupType()
	err = data.SetFile(ctx.File("cfgFile"))
	if err != nil {
		t.Fatalf("Error setting up config: %q", err)
	}
	list, err := data.List()
	if err != nil {
		t.Fatalf("Error getting list: %q", err)
	}

	expected := map[string]string{
		// "Name":         "",
		"AnInt": "0",
		"ABool": "false",
		// "Manager.Name": "",
		"Manager.Age": "0",
	}

	if err = CheckAllMap(list, expected); err != nil {
		t.Fatal(err)
	}

	_, err = os.Stat(data.File())
	if err == nil || !strings.Contains(err.Error(), "no such file") {
		t.Fatalf("Config file should not be on disk yet!")
	}

	pass()
}

func TestCreateEmptyConfig(t *testing.T) {
	ctx, err := NewContext(nil)
	if err != nil {
		t.Fatalf("Error creating context: %q", err)
	}
	defer ctx.Delete()

	data := NewGroupType()
	err = data.SetFile(ctx.File("cfgFile"))
	if err != nil {
		t.Fatalf("Error setting up config: %q", err)
	}

	_, err = os.Stat(data.File())
	if err == nil || !strings.Contains(err.Error(), "no such file") {
		t.Fatalf("Config file should not be on disk yet!")
	}

	// Save to disk
	if err = data.Save(); err != nil {
		t.Fatalf("Error saving: %q", err)
	}

	_, err = os.Stat(data.File())
	if err != nil {
		t.Fatalf("Config file should exist: ", err)
	}

	// Reload from scratch
	data = NewGroupType()
	err = data.SetFile(ctx.File("cfgFile"))
	if err != nil {
		t.Fatalf("Error reloading up config: %q", err)
	}
	list, err := data.List()
	if err != nil {
		t.Fatalf("Error getting list: %q", err)
	}

	expected := map[string]string{
		// "Name":         "",
		"AnInt": "0",
		"ABool": "false",
		// "Manager.Name": "",
		"Manager.Age": "0",
	}

	if err = CheckAllMap(list, expected); err != nil {
		t.Fatal(err)
	}

	pass()
}

func TestList(t *testing.T) {
	data := &GroupType{
		Name:  "Dept",
		AnInt: 123,
		ABool: true,
		Manager: PersonType{
			Name: "Joe",
			Age:  53,
		},
		StrArray: []string{
			"line4",
			"line9",
		},
		People: []PersonType{
			{Name: "P1", Age: 12},
			{Name: "P2", Age: 39},
		},
		PeopleMap: map[string]PersonType{
			"John": {
				Name: "Johnny",
				Age:  23,
			},
		},
	}
	data.MakeConfig(data)

	expected := map[string]string{
		"Name":                "Dept",
		"AnInt":               "123",
		"ABool":               "true",
		"Manager.Name":        "Joe",
		"Manager.Age":         "53",
		"StrArray.1":          "line4",
		"StrArray.2":          "line9",
		"People.1.Name":       "P1",
		"People.1.Age":        "12",
		"People.2.Name":       "P2",
		"People.2.Age":        "39",
		"PeopleMap.John.Name": "Johnny",
		"PeopleMap.John.Age":  "23",
	}

	list, err := data.List()
	if err != nil {
		t.Fatalf("Error getting list: %q", err)
	}

	if err = CheckAllMap(list, expected); err != nil {
		t.Fatal(err)
	}

	keys := data.Keys()
	for k1 := range expected {
		found := false
		for j, k2 := range keys {
			if k2 == k1 {
				keys = append(keys[:j], keys[j+1:]...)
				found = true
				break
			}
		}
		if !found {
			t.Fatalf("Can't find %s", k1)
		}
	}
	if len(keys) != 0 {
		t.Fatalf("Extra keys!: %q", keys)
	}

	pass()
}

func TestSimpleSet(t *testing.T) {
	ctx, err := NewContext(nil)
	if err != nil {
		t.Fatalf("Error creating context: %q", err)
	}
	defer ctx.Delete()

	data := &GroupType{
		Name:  "Dept",
		AnInt: 123,
		ABool: true,
		Manager: PersonType{
			Name: "Joe",
			Age:  53,
		},
		StrArray: []string{
			"line4",
			"line9",
		},
		People: []PersonType{
			{Name: "P1", Age: 12},
			{Name: "P2", Age: 39},
		},
		PeopleMap: map[string]PersonType{
			"John": {
				Name: "Johnny",
				Age:  23,
			},
		},
	}
	data.MakeConfig(data)

	err = data.SetFile(ctx.File("cfgFile"))
	if err != nil {
		t.Fatalf("Error setting up config: %q", err)
	}

	expected := map[string]string{
		"Name":                "Tepd",
		"AnInt":               "321",
		"ABool":               "false",
		"Manager.Name":        "eoJ",
		"Manager.Age":         "35",
		"StrArray.1":          "4enil",
		"StrArray.2":          "9enil",
		"People.1.Name":       "1P",
		"People.1.Age":        "21",
		"People.2.Name":       "2P",
		"People.2.Age":        "93",
		"PeopleMap.John.Name": "Ynnhoj",
		"PeopleMap.John.Age":  "32",
	}

	for k := range expected {
		Set(t, &data.Config, k, expected[k])
	}

	for i := 0; i < 2; i++ {
		Check(t, &data.Config, "Name", data.Name, "Tepd")
		Check(t, &data.Config, "AnInt", data.AnInt, 321)
		Check(t, &data.Config, "ABool", data.ABool, false)
		Check(t, &data.Config, "Manager.Name", data.Manager.Name, "eoJ")
		Check(t, &data.Config, "Manager.Age", data.Manager.Age, 35)
		Check(t, &data.Config, "StrArray.1", data.StrArray[0], "4enil")
		Check(t, &data.Config, "StrArray.2", data.StrArray[1], "9enil")
		Check(t, &data.Config, "People.1.Name", data.People[0].Name, "1P")
		Check(t, &data.Config, "People.1.Age", data.People[0].Age, 21)
		Check(t, &data.Config, "People.2.Name", data.People[1].Name, "2P")
		Check(t, &data.Config, "People.2.Age", data.People[1].Age, 93)
		Check(t, &data.Config, "PeopleMap.John.Name", data.PeopleMap["John"].Name, "Ynnhoj")

		// Now check List/Get too
		list, err := data.List()
		if err != nil {
			t.Fatalf("Error getting list: %q", err)
		}

		if err = CheckAllMap(list, expected); err != nil {
			t.Fatal(err)
		}

		// Now save it, reload and check it all again
		data.Save()
		data = NewGroupType()
		err = data.SetFile(ctx.File("cfgFile"))
		if err != nil {
			t.Fatalf("Error setting up config: %q", err)
		}
	}

	pass()
}

func TestStringArray(t *testing.T) {
	ctx, err := NewContext(nil)
	if err != nil {
		t.Fatalf("Error creating context: %q", err)
	}
	defer ctx.Delete()

	data := NewGroupType()
	err = data.SetFile(ctx.File("cfgFile"))
	if err != nil {
		t.Fatalf("Error setting up config: %q", err)
	}

	_, err = data.Get("StrArray.1")
	if err == nil {
		t.Fatalf("Should not have been able to get StrArray.1")
	}

	Set(t, &data.Config, "StrArray.1", "Jenny")
	Check(t, &data.Config, "StrArray.1", data.StrArray[0], "Jenny")

	Set(t, &data.Config, "StrArray.1", "")
	_, err = data.Get("StrArray.1")
	if err == nil {
		t.Fatalf("Should not have been able to get StrArray.1")
	}

	Set(t, &data.Config, "StrArray.2", "Jenny")
	Check(t, &data.Config, "StrArray.1", data.StrArray[0], "")
	Check(t, &data.Config, "StrArray.2", data.StrArray[1], "Jenny")

	pass()
}

func TestString(t *testing.T) {
	ctx, err := NewContext(nil)
	if err != nil {
		t.Fatalf("Error creating context: %q", err)
	}
	defer ctx.Delete()

	data := NewGroupType()
	err = data.SetFile(ctx.File("cfgFile"))
	if err != nil {
		t.Fatalf("Error setting up config: %q", err)
	}

	exp := "John"
	data.Name = exp
	Check(t, &data.Config, "Name", data.Name, exp)

	if err = data.Save(); err != nil {
		t.Fatalf("Error saving: %q", err)
	}

	data = NewGroupType()
	err = data.SetFile(ctx.File("cfgFile"))
	if err != nil {
		t.Fatalf("Error reloading up config: %q", err)
	}

	Check(t, &data.Config, "Name", data.Name, exp)

	exp = "George"
	Set(t, &data.Config, "Name", exp)
	Check(t, &data.Config, "Name", data.Name, exp)

	pass()
}

func TestInt(t *testing.T) {
	ctx, err := NewContext(nil)
	if err != nil {
		t.Fatalf("Error creating context: %q", err)
	}
	defer ctx.Delete()

	data := NewGroupType()
	err = data.SetFile(ctx.File("cfgFile"))
	if err != nil {
		t.Fatalf("Error setting up config: %q", err)
	}

	exp := 99
	data.AnInt = exp
	Check(t, &data.Config, "AnInt", data.AnInt, exp)

	err = data.Save()
	if err != nil {
		t.Fatalf("Error saving: %q", err)
	}

	data = NewGroupType()
	err = data.SetFile(ctx.File("cfgFile"))
	Check(t, &data.Config, "AnInt", data.AnInt, exp)

	exp = 666
	Set(t, &data.Config, "AnInt", exp)
	Check(t, &data.Config, "AnInt", data.AnInt, exp)

	pass()
}

func TestSaveLoad(t *testing.T) {
	ctx, err := NewContext(nil)
	if err != nil {
		t.Fatalf("Error creating context: %q", err)
	}
	defer ctx.Delete()

	data := &GroupType{
		Name:  "John",
		AnInt: 77,
		ABool: true,
		Manager: PersonType{
			Name: "Mary",
			Age:  21,
		},
		StrArray: []string{
			"line1",
			"line2",
		},
		PeopleMap: map[string]PersonType{
			"John": {"Johnny", 23},
		},
	}
	data.MakeConfig(data)
	err = data.SetFile(ctx.File("cfgFile"))
	if err != nil {
		t.Fatalf("Error setting up config: %q", err)
	}

	err = data.Save()
	if err != nil {
		t.Fatalf("Error saving: %q", err)
	}

	// Now reload it
	data = NewGroupType()
	err = data.SetFile(ctx.File("cfgFile"))
	if err != nil {
		t.Fatalf("Error reloading up config: %q", err)
	}

	expected := map[string]string{
		"Name":                "John",
		"AnInt":               "77",
		"ABool":               "true",
		"Manager.Name":        "Mary",
		"Manager.Age":         "21",
		"StrArray.1":          "line1",
		"StrArray.2":          "line2",
		"PeopleMap.John.Name": "Johnny",
		"PeopleMap.John.Age":  "23",
	}

	list, err := data.List()
	if err != nil {
		t.Fatalf("Error getting list: %q", err)
	}

	if err = CheckAllMap(list, expected); err != nil {
		t.Fatal(err)
	}

	pass()
}

func TestSliceString(t *testing.T) {
	ctx, err := NewContext(nil)
	if err != nil {
		t.Fatalf("Error creating context: %q", err)
	}
	defer ctx.Delete()

	data := NewGroupType()
	err = data.SetFile(ctx.File("cfgFile"))
	if err != nil {
		t.Fatalf("Error setting up config: %q", err)
	}

	data.StrArray = []string{"line1", "line2"}
	Check(t, &data.Config, "StrArray.1", data.StrArray[0], "line1")
	Check(t, &data.Config, "StrArray.2", data.StrArray[1], "line2")

	err = data.Set("StrArray.1", "new line1")
	if err != nil {
		t.Fatalf("Error setting StrArray.1: %q", err)
	}
	err = CheckSlice(data.StrArray, []string{"new line1", "line2"})
	if err != nil {
		t.Fatalf("data.StrArray test 1: %q", err)
	}

	err = data.Set("StrArray.1", "")
	if err != nil {
		t.Fatalf("Error setting StrArray.1: %q", err)
	}
	err = CheckSlice(data.StrArray, []string{"line2"})
	if err != nil {
		t.Fatalf("data.StrArray test 2: %q", err)
	}

	err = data.Set("StrArray.1", "")
	if err != nil {
		t.Fatalf("Error setting StrArray.1: %q", err)
	}
	err = CheckSlice(data.StrArray, []string{})
	if err != nil {
		t.Fatalf("data.StrArray test 3: %q", err)
	}

	err = data.Set("StrArray.1", "one")
	if err != nil {
		t.Fatalf("Error setting StrArray.1: %q", err)
	}
	err = CheckSlice(data.StrArray, []string{"one"})
	if err != nil {
		t.Fatalf("data.StrArray test 4: %q", err)
	}

	err = data.Set("StrArray.2", "two")
	if err != nil {
		t.Fatalf("Error setting StrArray.2: %q", err)
	}
	err = CheckSlice(data.StrArray, []string{"one", "two"})
	if err != nil {
		t.Fatalf("data.StrArray test 5: %q", err)
	}

	err = data.Set("StrArray.4", "four")
	if err != nil {
		t.Fatalf("Error setting StrArray.4: %q", err)
	}
	err = CheckSlice(data.StrArray, []string{"one", "two", "", "four"})
	if err != nil {
		t.Fatalf("data.StrArray test 6: %q", err)
	}

	pass()
}

func TestMapStruct(t *testing.T) {
	ctx, err := NewContext(nil)
	if err != nil {
		t.Fatalf("Error creating context: %q", err)
	}
	defer ctx.Delete()

	data := &GroupType{
		PeopleMap: map[string]PersonType{
			"John": {Name: "Johnny", Age: 66},
		},
	}
	data.MakeConfig(data)
	err = data.SetFile(ctx.File("cfgFile"))
	if err != nil {
		t.Fatalf("Error setting up config: %q", err)
	}

	val, err := data.Get("PeopleMap.John.Name")
	if err != nil {
		t.Fatalf("Error getting PeopleMap.John.Name: %q", err)
	}
	if val != "Johnny" {
		t.Fatalf("Wrong value for PeopleMap.John.Name: %s", val)
	}

	err = data.Set("PeopleMap.John.Age", "23")
	if err != nil {
		t.Fatalf("Error setting PeopleMap.John.Name: %q", err)
	}

	err = data.Set("PeopleMap.John", "")
	if err != nil || len(data.PeopleMap) != 0 {
		t.Fatalf("Error deleting PeopleMap.John:%q\nLen: %d", err, len(data.PeopleMap))
	}

	err = data.Set("PeopleMap.Mary.Name", "Marie")
	if err != nil {
		t.Fatalf("Error adding Mary: %q\n", err)
	}
	err = data.Set("PeopleMap.Mary.Age", "99")
	if err != nil || data.PeopleMap["Mary"].Age != 99 || data.PeopleMap["Mary"].Name != "Marie" {
		t.Fatalf("Error updating Mary - err:%q\ndata: %q", err, data)
	}

	pass()
}

func TestSliceStruct(t *testing.T) {
	ctx, err := NewContext(nil)
	if err != nil {
		t.Fatalf("Error creating context: %q", err)
	}
	defer ctx.Delete()

	data := NewGroupType()
	err = data.SetFile(ctx.File("cfgFile"))
	if err != nil {
		t.Fatalf("Error setting up config: %q", err)
	}

	data.People = []PersonType{{"John", 5}, {"Mary", 10}}

	Check(t, &data.Config, "People.1.Name", data.People[0].Name, "John")
	Check(t, &data.Config, "People.2.Age", data.People[1].Age, 10)

	Set(t, &data.Config, "People.1.Name", "Steve")
	Check(t, &data.Config, "People.1.Name", data.People[0].Name, "Steve")

	// Erase 2nd one
	Set(t, &data.Config, "People.2", "")

	Set(t, &data.Config, "People.2.Age", "42")
	Check(t, &data.Config, "People.2.Name", data.People[1].Name, "")
	Check(t, &data.Config, "People.2.Age", data.People[1].Age, 42)

	pass()
}

func TestLoad(t *testing.T) {
	ctx, err := NewContext(map[string]string{
		"badData":  "not json!",
		"goodData": `{"Name": "John"}`,
	})
	if err != nil {
		t.Fatalf("Error creating context: %q", err)
	}
	defer ctx.Delete()

	data := NewGroupType()

	err = data.Load()
	if err == nil || !strings.Contains(err.Error(), "No file defined") {
		t.Fatalf("Should have failed about no file defined: %q", err)
	}

	data = NewGroupType()
	err = data.SetFile(ctx.File("Foo"))
	if err != nil {
		t.Fatalf("Error setting up config: %q", err)
	}

	err = data.Load()
	if err == nil || !strings.Contains(err.Error(), "no such file") {
		t.Fatalf("Should have failed about no file defined: %q", err)
	}

	data = NewGroupType()
	err = data.SetFile(ctx.File("badData"))
	if err == nil || !strings.Contains(err.Error(), "invalid character") {
		t.Fatalf("Should have failed to load data: %q", err)
	}

	// Mess with file and then try to reload it
	data = NewGroupType()
	err = data.SetFile(ctx.File("Missing"))
	if err != nil {
		t.Fatalf("Error setting up config: %q", err)
	}

	err = ioutil.WriteFile(ctx.File("Missing"), []byte("foo"), 0700)
	if err != nil {
		t.Fatalf("Error updating file: %q", err)
	}
	err = data.Load()
	if err == nil || !strings.Contains(err.Error(), "invalid character") {
		t.Fatalf("Should have failed about no file defined: %q", err)
	}

	// Now a good case
	data = NewGroupType()
	err = data.SetFile(ctx.File("goodData"))
	if err != nil {
		t.Fatalf("Error loading file: %q", err)
	}
	Check(t, &data.Config, "Name", data.Name, "John")

	pass()
}

func TestErrorSave(t *testing.T) {
	ctx, err := NewContext(map[string]string{
		"file1": "{}",
	})
	if err != nil {
		t.Fatalf("Error creating context: %q", err)
	}
	defer ctx.Delete()

	data := NewGroupType()

	err = data.Save()
	if err == nil || !strings.Contains(err.Error(), "Missing file") {
		t.Fatalf("Should have failed to save (no file): %q", err)
	}

	pass()
}

func TestErrorJsonFile(t *testing.T) {
	fName := "file"

	ctx, err := NewContext(map[string]string{
		fName: "{}",
	})
	if err != nil {
		t.Fatalf("Error creating context: %q", err)
	}
	defer ctx.Delete()

	data := NewGroupType()
	err = data.SetFile(ctx.File(fName))
	if err != nil {
		t.Fatalf("Error setting up config: %q", err)
	}

	if data.File() != ctx.File(fName) {
		t.Fatalf("Wrong filename: %s", data.File())
	}

	pass()
}

func TestErrorBadSelector(t *testing.T) {
	ctx, err := NewContext(map[string]string{})
	if err != nil {
		t.Fatalf("Error creating context: %q", err)
	}
	defer ctx.Delete()

	data := NewGroupType()
	err = data.SetFile(ctx.File("cfgFile"))
	if err != nil {
		t.Fatalf("Error setting up config: %q", err)
	}

	_, err = data.Get("")
	if err == nil || !strings.Contains(err.Error(), "Missing key") {
		t.Fatalf("Finding '' should have failed: %q", err)
	}

	_, err = data.Get("Foo")
	if err == nil || !strings.Contains(err.Error(), "No field with name") {
		t.Fatalf("Finding 'Foo' should have failed: %q", err)
	}

	_, err = data.Get("Name.Foo")
	if err == nil || !strings.Contains(err.Error(), "Can't step into") {
		t.Fatalf("Finding 'Foo' should have failed: %q", err)
	}

	pass()
}

func TestErrorBadType(t *testing.T) {
	ctx, err := NewContext(map[string]string{})
	if err != nil {
		t.Fatalf("Error creating context: %q", err)
	}
	defer ctx.Delete()

	data := &struct {
		JsonConfig
		Data func()
	}{}
	data.MakeConfig(data)
	err = data.SetFile(ctx.File("cfgFile"))
	if err != nil {
		t.Fatalf("Error setting up config: %q", err)
	}

	_, err = data.Get("Data")
	if err == nil || !strings.Contains(err.Error(), "Unsupported") {
		t.Fatalf("Finding 'Data' should have failed: %q", err)
	}

	pass()
}

func TestErrorBadSliceIndex(t *testing.T) {
	ctx, err := NewContext(map[string]string{})
	if err != nil {
		t.Fatalf("Error creating context: %q", err)
	}
	defer ctx.Delete()

	data := &GroupType{
		StrArray: []string{"line1", "line2"},
	}
	data.MakeConfig(data)
	err = data.SetFile(ctx.File("cfgFile"))
	if err != nil {
		t.Fatalf("Error setting up config: %q", err)
	}

	if _, err = data.Get("StrArray"); err == nil {
		t.Fatalf("Can't return slice")
	}

	_, err = data.Get("StrArray.3")
	if err == nil || !strings.Contains(err.Error(), "out of range") {
		t.Fatalf("Getting StrArray.3 should have failed: %q", err)
	}

	_, err = data.Get("StrArray.Foo")
	if err == nil || !strings.Contains(err.Error(), "converting") {
		t.Fatalf("Getting StrArray.Foo should have failed")
	}

	pass()
}

func TestErrorReturnStruct(t *testing.T) {
	ctx, err := NewContext(map[string]string{})
	if err != nil {
		t.Fatalf("Error creating context: %q", err)
	}
	defer ctx.Delete()

	data := &GroupType{
		StrArray: []string{"line1", "line2"},
	}
	data.MakeConfig(data)
	err = data.SetFile(ctx.File("cfgFile"))
	if err != nil {
		t.Fatalf("Error setting up config: %q", err)
	}

	_, err = data.Get("Manager")
	if err == nil || !strings.Contains(err.Error(), "Unsupported") {
		t.Fatalf("Getting 'Manager' should have failed: %q", err)
	}

	pass()
}

func TestErrorSetKeys(t *testing.T) {
	data := NewGroupType()

	err := data.Set("", "")
	if err == nil || !strings.Contains(err.Error(), "Missing key") {
		t.Fatalf("Should have failed on Set(''): %q", err)
	}

	err = data.Set("Name", "foo")
	if err != nil {
		t.Fatalf("Setting Name should have worked: %q", err)
	}

	err = data.Set("Name.1", "foo")
	if err == nil || !strings.Contains(err.Error(), "step into") {
		t.Fatalf("Should have failed on Set(''): %q", err)
	}

	err = data.Set("AnInt", "5")
	if err != nil {
		t.Fatalf("Setting AnInt should have worked: %q", err)
	}

	err = data.Set("AnInt.1", "foo")
	if err == nil || !strings.Contains(err.Error(), "step into") {
		t.Fatalf("Should have failed on Set(''): %q", err)
	}

	err = data.Set("AnInt", "foo")
	if err == nil || !strings.Contains(err.Error(), "invalid syntax") {
		t.Fatalf("Should have failed on AnInt(foo): %q", err)
	}

	err = data.Set("ABool", "true")
	if err != nil {
		t.Fatalf("Setting ABool should have worked: %q", err)
	}

	err = data.Set("ABool.1", "true")
	if err == nil || !strings.Contains(err.Error(), "step into") {
		t.Fatalf("Setting ABool.1 should have failed: %q", err)
	}

	err = data.Set("ABool", "boogie")
	if err != nil || data.ABool != false {
		t.Fatalf("Setting ABool='boogie' should have worked and it should be 'true': err:%q\nvalue:%v", err, data.ABool)
	}

	pass()
}

func TestDump(t *testing.T) {
	ctx, err := NewContext(nil)
	if err != nil {
		t.Fatalf("Error creating context: %q", err)
	}
	defer ctx.Delete()

	/*  TODO - re-enable
	data := &struct{ JsonConfig }{}
	data.MakeConfig(data)

	str, err := data.Dump()
	if err != nil {
		t.Fatalf("Dump failed: %q", err)
	}

	if str != "{}" {
		t.Fatalf("Wrong output from Dump: %s", str)
	}
	*/

	data2 := &struct {
		JsonConfig
		A string
	}{}
	data2.MakeConfig(data2)

	str, err := data2.Dump()
	if err != nil {
		t.Fatalf("Dump failed: %q", err)
	}

	if str != "{\n  \"A\": \"\"\n}" {
		t.Fatalf("Wrong output from Dump({}): %q", str)
	}

	pass()
}

func TestIntMapKeys(t *testing.T) {
	type myStruct struct {
		JsonConfig
		MyMap map[int]string
	}
	myConfig := &myStruct{
		MyMap: map[int]string{
			3: "three",
			6: "six",
		},
	}
	myConfig.MakeConfig(myConfig)

	Check(t, &myConfig.Config, "MyMap.3", myConfig.MyMap[3], "three")

	err := myConfig.Set("MyMap.3", "eerht")
	if err != nil {
		t.Fatalf("Error from set(eerht): %q", err)
	}
	Check(t, &myConfig.Config, "MyMap.3", myConfig.MyMap[3], "eerht")

	err = myConfig.Set("MyMap.3", "")
	if err != nil {
		t.Fatalf("Error from set(''): %q", err)
	}

	_, err = myConfig.Get("MyMap.3")
	if err == nil || !strings.Contains(err.Error(), "No entry") {
		t.Fatalf("Bad result from Get(MyMap.3): %q", err)
	}

	myConfig.Set("MyMap.7", "seven")
	Check(t, &myConfig.Config, "MyMap.7", myConfig.MyMap[7], "seven")

	pass()
}

func TestAnonStruct(t *testing.T) {
	type Inner struct {
		JsonConfig
		A int
	}
	type myStruct struct {
		Inner
		B int
	}
	myConfig := &myStruct{}
	myConfig.MakeConfig(myConfig)

	myConfig.Set("A", "11")
	myConfig.Set("B", "99")

	Check(t, &myConfig.Config, "A", myConfig.A, 11)
	Check(t, &myConfig.Config, "B", myConfig.B, 99)

	expected := map[string]string{
		"B": "99",
		"A": "11",
	}

	list, err := myConfig.List()
	if err != nil {
		t.Fatalf("Error getting list: %q", err)
	}

	if err = CheckAllMap(list, expected); err != nil {
		t.Fatal(err)
	}

	pass()
}
