// testing that reverse order works ok
===
===
{
  "bool1": true,
  "bool2": false,
  "float1": 1.0,
  "string1": "hello world",
  "array1": [
    { "elem1": "e1"},
	{ "elem2": "e2"}
  ],
  "map1": {
    "m1": "m1value",
	"m2": "m2value"
  },
  "null1": null
}
===
{    
  "null1": null,
  "map1": {
    "m1": "m1value",
	"m2": "m2value"
  },
  "array1": [
    {"elem1": "e1"},
	{"elem2": "e2"}
  ],
  "string1": "hello world",
  "float1": 1.0,
  "bool2": false,
  "bool1": true
}
