Configuration File Manager
==================================

This package will manage configuration files. While people normally will
simply save their data, as json, to a file and (un)marshal it back and
forth, this package will also allow you to dynamically manipulate the
data via simple strings.  For example, let's say you have
a config struct that looks like:

```
type Person struct {
	Name string
	Age	 int
}

myData := &Person{}
```

While you can do:
```
myData.Name = "John"
```
without this package, you can't easily do:
```
myData.Set("Name", "John")
```
which would be needed if you wanted to allow someone to manipulate
the data via a simple command line interface.  For example:
```
myApp set Name John
```
Its this dynamic processing that drove the need for this package.

The following features are supported:
* `Get( fieldName string )` - Retrieve the value of `fieldName`.
* `Set( fieldName string, fieldValue string)` - Update the value of `fieldName`.
* `List()` - Show the values of all fields in the config resource.
* `Keys()` - Show all of the possible fieldNames.
* `Dump()` - Show the raw json version of the config data.
* `Load()` - Load the config data from a file.
* `Save()` - Save the config data to a file.

The format of `fieldName` is a dot(.) separated list of field names
within the structures.

Arrays elements are specified by specifying the index you want to 
get/set. For example:
```
type Person struct {
	AList []string
	Name string
	Age	 int
}

myData := &Person{}
myData.Set( "AList.2", "John" )
```

