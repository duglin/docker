package jsondiff

// This package will diff two JSON files.
// Its not really meant to be very thorough - meaning, its not too smart
// w.r.t. finding matches so the output isn't quite a condensed as it could be.
// But the point here isn't to create 'ed' type of output, its more to
// return a boolean result as to whether or not something has changed, with
// a rough pointer for where to look for the diff.
// With that said, it should at least not lie about whether there are diffs
// or not.  :-)

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"reflect"
	"sort"
)

// extraOK indicates if its ok for the 2nd file we're diffing to have
// additional fields that are not present in the 1st file
var extraOK bool

// typeOf just returns a string representation of the golang type of 'v'.
func typeOf(v interface{}) string {
	t := reflect.TypeOf(v)

	if t == nil {
		return "null"
	}
	s := t.String()
	if s == "map[string]interface {}" {
		s = "object"
	} else if s == "[]interface {}" {
		s = "array"
	}
	return s
}

// OneLiner will serialize 'v' into JSON w/o any newlines between the fields
func OneLiner(v interface{}) string {
	res, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	return string(res)
}

// diffJson will compare two two JSON decoded objects, returning
// the "diff" as a string.
func diffJson(j1, j2 interface{}, parent string) string {
	diff := ""
	didParent := false

	t1 := typeOf(j1)
	t2 := typeOf(j2)

	// If the two types aren't the same then we can just stop immediately
	if t1 != t2 {
		if parent != "" {
			diff += fmt.Sprintf("%s: (type)\n", parent)
		}

		diff += fmt.Sprintf("< %s\n---\n> %s\n", t1, t2)
		return diff
	}

	switch t1 {
	case "object":
		// JSON Objects are just maps so make sure that each "key" in
		// in the 1st appears in the 2nd and then diff values.
		m1, _ := j1.(map[string]interface{})
		m2, _ := j2.(map[string]interface{})
		seenList := map[string]bool{}

		// For consistency, sort the keys
		keys1 := []string{}
		for k1 := range m1 {
			keys1 = append(keys1, k1)
		}
		sort.Strings(keys1)

		keys2 := []string{}
		for k2 := range m2 {
			keys2 = append(keys2, k2)
		}
		sort.Strings(keys1)

		for _, k1 := range keys1 {
			v1 := m1[k1]
			v2, ok := m2[k1]
			if !ok {
				/*
					if !didParent && parent != "" {
						didParent = true
						diff += fmt.Sprintf("%s:\n", parent)
					}
				*/
				tmpParent := parent
				if parent != "" {
					tmpParent += "."
				}
				// diff += fmt.Sprintf("< %s\n", k1)
				diff += fmt.Sprintf("- %s%s\n", tmpParent, k1)
				continue
			}
			// Keep track of which keys in the 2nd file we're seen
			seenList[k1] = true
			tmpParent := parent
			if parent != "" {
				tmpParent += "."
			}
			diff += diffJson(v1, v2, tmpParent+k1)
		}

		// Any keys in the 2nd file we haven't seen are treated as
		// new/extra fields.
		if !extraOK {
			for _, k2 := range keys2 {
				if _, ok := seenList[k2]; ok {
					continue
				}
				_, ok := m1[k2]
				if !ok {
					/*
						if !didParent && parent != "" {
							didParent = true
							diff += fmt.Sprintf("%s:\n", parent)
						}
					*/
					tmpParent := parent
					if parent != "" {
						tmpParent += "."
					}
					// diff += fmt.Sprintf("> %s\n", k2)
					diff += fmt.Sprintf("+ %s%s\n", tmpParent, k2)
					continue
				}
			}
		}

	case "array":
		// Very simplistic diff here.  We just loop over the first array
		// and see where it is in the 2nd. We treat the value of each
		// entry in its entirety, we don't really both to do smart diffing
		// all the way down the tree to show just the inner-most bit that
		// changed. That would be good if we ever have deep trees but for now
		// I think we can get by with this dumb algorithm.
		// If someone is bored, this would be the section to "smart up".
		arr1, _ := j1.([]interface{})
		arr2, _ := j2.([]interface{})

		// Where in file 2 we stopped search last time
		pos2 := 0

		// For each array item in file 1 ...
		for pos1, v1 := range arr1 {
			hit := false

			// Look for it in file 2, from where we left off last time (pos2)
			for i := pos2; i < len(arr2); i++ {
				v2 := arr2[i]
				tmpDiff := diffJson(v1, v2, parent)

				// Found a match so anything between this and the previous
				// spot (pos2) is new data in file 2
				if tmpDiff == "" {
					if !extraOK {
						for j := pos2; j < i; j++ {
							v2 = arr2[j]
							if !didParent && parent != "" {
								didParent = true
								diff += fmt.Sprintf("%s:\n", parent)
							}
							diff += fmt.Sprintf("> [%d] %.50s\n", j+1, OneLiner(v2))
						}
					}
					pos2 = i + 1
					hit = true
					break
				}
			}
			if hit {
				// Found a match in file 2 so move on to the rest one in file 1
				continue
			}

			// Didn't find it in file 2 so it must be new in file 1
			if !didParent && parent != "" {
				didParent = true
				diff += fmt.Sprintf("%s:\n", parent)
			}
			// diff += fmt.Sprintf("< [%d] %.50s\n", pos1+1, OneLiner(v1))
			diff += fmt.Sprintf("- [%d] %.50s\n", pos1+1, OneLiner(v1))
		}

		// Anything left over in file 2 is considered new
		if !extraOK && pos2 != len(arr2) {
			if !didParent && parent != "" {
				didParent = true
				diff += fmt.Sprintf("%s:\n", parent)
			}
			for i := pos2; i < len(arr2); i++ {
				v2 := arr2[i]
				// diff += fmt.Sprintf("> [%d] %.50s\n", i+1, OneLiner(v2))
				diff += fmt.Sprintf("+ [%d] %.50s\n", i+1, OneLiner(v2))
			}
		}

	default:
		// Simple data types are easy
		if j1 != j2 {
			if parent != "" {
				diff += fmt.Sprintf("%s:\n", parent)
			}
			diff += fmt.Sprintf("< %v\n---\n> %v\n", j1, j2)
		}

	}

	return diff
}

// diffFiles will compare two JSON files and return the diff as a string
func diffFiles(file1, file2 string) (string, error) {
	j1 := map[string]interface{}{}
	j2 := map[string]interface{}{}

	f1, err := os.Open(file1)
	if err != nil {
		return "", fmt.Errorf("Error opening file '%s': %s", file1, err)
	}
	defer f1.Close()
	dec1 := json.NewDecoder(f1)
	if err := dec1.Decode(&j1); err != nil {
		return "", fmt.Errorf("Error parsing '%s': %s", file1, err)
	}
	f1.Close()

	f2, err := os.Open(file2)
	if err != nil {
		return "", fmt.Errorf("Error opening file '%s': %s", file2, err)
	}
	defer f2.Close()
	dec2 := json.NewDecoder(f2)
	if err := dec2.Decode(&j2); err != nil {
		return "", fmt.Errorf("Error parsing '%s': %s", file2, err)
	}
	f2.Close()

	return diffJson(j1, j2, ""), nil
}

func main() {
	flag.BoolVar(&extraOK, "e", false, "Extra data in 2nd file is ok")
	flag.Parse()

	if flag.NArg() != 2 {
		fmt.Fprintf(os.Stderr, "Requires two arguments/files\n")
		os.Exit(1)
	}

	diff, err := diffFiles(flag.Args()[0], flag.Args()[1])

	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
	}
	if diff != "" {
		fmt.Fprintf(os.Stderr, "%s", diff)
		os.Exit(1)
	}
	os.Exit(0)
}
