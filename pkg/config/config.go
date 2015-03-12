package config

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"unicode"
)

// Generic Config Stuff
type Config struct {
	data interface{}
}

func (cfg *Config) MakeConfig(d interface{}) {
	cfg.data = d
}

func (cfg *Config) Get(key string) (string, error) {
	keys := strings.Split(strings.TrimSpace(key), ".")
	field := reflect.ValueOf(cfg.data).Elem()

	if key == "" {
		return "", fmt.Errorf("Missing key")
	}

	for i, key := range keys {
		/*
			if unicode.IsLower(rune(key[0])) {
				return "", fmt.Errorf("Can't specify unexported field: %s", key)
			}
		*/

		switch field.Kind() {
		case reflect.Map:
			keyType := field.Type().Key().Kind()
			var keyValue reflect.Value
			if keyType == reflect.String {
				keyValue = reflect.ValueOf(key)
			} else if keyType == reflect.Int {
				i, err := strconv.Atoi(key)
				if err != nil {
					return "", err
				}
				keyValue = reflect.ValueOf(i)
			} else {
				return "", fmt.Errorf("Can't support map keys of type: %q", keyType)
			}

			field = field.MapIndex(keyValue)
			if !field.IsValid() {
				return "", fmt.Errorf("No entry found with key: %q", key)
			}

		case reflect.Struct:
			field = field.FieldByName(key)
			if !field.IsValid() {
				return "", fmt.Errorf("No field with name %q", key)
			}

		case reflect.Slice:
			index, err := strconv.Atoi(keys[i])
			if err != nil {
				return "", fmt.Errorf("Error converting %q to an int", keys[i])
			}
			if index < 1 || index > field.Len() {
				return "", fmt.Errorf("Index (%s) is out of range", keys[i])
			}
			field = field.Index(index - 1)

		default:
			return "", fmt.Errorf("Can't step into a %q via %q", field.Kind(), key)
		}

		if !field.IsValid() {
			return "", fmt.Errorf("Field is invalid: %s\n", key)
		}
	}

	k := field.Kind()
	if k != reflect.Int && k != reflect.String && k != reflect.Bool {
		return "", fmt.Errorf("Unsupported return type: %q", k)
	}

	return fmt.Sprintf("%v", field.Interface()), nil
}

type MapSetter struct {
	daMap      reflect.Value
	daMapKey   reflect.Value
	daMapEntry reflect.Value
}

func (cfg *Config) Set(key string, val string) error {
	key = strings.TrimSpace(key)
	if key == "" {
		return fmt.Errorf("Missing key")
	}

	keys := strings.Split(strings.TrimSpace(key), ".")
	field := reflect.ValueOf(cfg.data).Elem()

	var prevWord string

	maps := []MapSetter{}

	for w := 0; w <= len(keys); w++ {
		var word string
		if w < len(keys) {
			word = keys[w]
		}

		/*
			if word != "" && unicode.IsLower(rune(word[0])) {
				return fmt.Errorf("Can't specify unexported field: %s", key)
			}
		*/

		// fmt.Printf("---\n")
		// fmt.Printf("field: %q\n", field)
		// fmt.Printf("word: %q\n", word)
		// fmt.Printf("kind: %q\n", field.Kind())
		// fmt.Printf("val: %q\n", val)

		switch field.Kind() {
		// TODO add bool
		case reflect.Int:
			if word != "" {
				return fmt.Errorf("Can't step into %s via %s", prevWord, keys[w:])
			}
			v, err := strconv.Atoi(val)
			if err != nil {
				return err
			}
			if !field.CanSet() {
				return fmt.Errorf("Can't set value: %s", prevWord)
			}
			field.SetInt(int64(v))

		case reflect.String:
			if word != "" {
				return fmt.Errorf("Can't step into %s via %s", prevWord, keys[w:])
			}
			if !field.CanSet() {
				return fmt.Errorf("Can't set value: %s", prevWord)
			}
			field.SetString(val)

		case reflect.Bool:
			if word != "" {
				return fmt.Errorf("Can't step into %s via %s", prevWord, keys[w:])
			}
			if !field.CanSet() {
				return fmt.Errorf("Can't set value: %s", prevWord)
			}
			field.SetBool(strings.ToLower(val) == "true")

		case reflect.Struct:
			field = field.FieldByName(word)
			if !field.IsValid() {
				return fmt.Errorf("No field with name %q", word)
			}

		case reflect.Map:
			keyType := field.Type().Key().Kind()
			var keyValue reflect.Value
			if keyType == reflect.String {
				keyValue = reflect.ValueOf(word)
			} else if keyType == reflect.Int {
				i, err := strconv.Atoi(word)
				if err != nil {
					return err
				}
				keyValue = reflect.ValueOf(i)
			} else {
				return fmt.Errorf("Can't support map keys of type: %q", keyType)
			}

			entry := field.MapIndex(keyValue)

			var dup reflect.Value
			if entry.IsValid() {
				// Found it
				if w+1 == len(keys) && val == "" {
					// Deleting it!
					field.SetMapIndex(keyValue, reflect.ValueOf(nil))
					return nil
				}

				// Dup it
				dup = CopyValue(entry)
			} else {
				// Not found - create a new one
				dup = reflect.New(field.Type().Elem()).Elem()
			}

			// Save info so we can set the entry in the map later after
			// we're all done
			ms := MapSetter{
				daMap:      field,
				daMapKey:   keyValue,
				daMapEntry: dup,
			}
			maps = append(maps, ms)

			field = dup

		case reflect.Slice:
			index, err := strconv.Atoi(word)
			if err != nil {
				return fmt.Errorf("Error converting %q to an int", word)
			}

			if index < 1 {
				return fmt.Errorf("Index (%s) is out of range", word)
			}

			fieldLen := field.Len()

			// fmt.Printf("Slice type: %q\n", field.Type())
			// fmt.Printf("Slice len: %q\n", fieldLen )
			// fmt.Printf("index: %d\n", index )

			if index > fieldLen {
				array := reflect.MakeSlice(field.Type(), index, index)
				for j := 0; j < fieldLen; j++ {
					array.Index(j).Set(field.Index(j))
				}
				field.Set(array)
				fieldLen = index
			}

			// If we're at the end of the keys then we can check to see
			// if we're erasing the entry
			if w+1 == len(keys) && val == "" {
				array := reflect.MakeSlice(field.Type(), fieldLen-1, fieldLen-1)
				count := 0
				for j := 0; j < fieldLen; j++ {
					if j == index-1 {
						continue
					}
					array.Index(count).Set(field.Index(j))
					count++
				}
				field.Set(array)

				// Delete is special and doesn't exit normally
				return nil
			} else {
				field = field.Index(index - 1)
			}

		default:
			return fmt.Errorf("Unsupported field type: %q", field.Kind())
		}

		prevWord = word
	}

	for i := len(maps) - 1; i >= 0; i-- {
		ms := maps[i]
		ms.daMap.SetMapIndex(ms.daMapKey, ms.daMapEntry)
	}

	return nil
}

func (cfg *Config) List() (map[string]string, error) {
	keys := cfg.Keys()
	result := map[string]string{}
	for _, key := range keys {
		val, err := cfg.Get(key)
		if err != nil {
			return nil, err
		}
		// Only show non-empty values
		if val != "" {
			result[key] = val
		}
	}
	return result, nil
}

func (cfg *Config) Dump() (string, error) {
	data, err := json.MarshalIndent(cfg.data, "", "  ")
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func (cfg *Config) Keys() []string {
	return keys(interface{}(cfg.data))
}

// keys returns the list of valid keys for the current struct.
// nil means the current obj is a simple type(int, string, etc...)
// empty []string means the current obj is not-simple but has no new keys
func keys(obj interface{}) []string {
	list := []string{}

	var field reflect.Value

	if reflect.ValueOf(obj).Kind() == reflect.Ptr {
		field = reflect.ValueOf(obj).Elem()
	} else {
		field = reflect.ValueOf(obj)
	}

	switch field.Kind() {

	case reflect.Map:
		daKeys := field.MapKeys()
		for _, k := range daKeys {
			newList := keys(field.MapIndex(k).Interface())
			if newList == nil { // || len(newList) == 0 {
				nn := fmt.Sprintf("%s", k.Interface())
				list = append(list, nn)
			} else {
				for _, val := range newList {
					nn := fmt.Sprintf("%v.%s", k.Interface(), val)
					list = append(list, nn)
				}
			}
		}

	case reflect.Struct:
		for i := 0; i < field.NumField(); i++ {
			valueField := field.Field(i)
			typeField := field.Type().Field(i)

			if unicode.IsLower(rune(typeField.Name[0])) {
				continue
			}

			newList := keys(valueField.Interface())
			if newList == nil { // || len(newList) == 0 {
				if !typeField.Anonymous {
					list = append(list, typeField.Name)
				}
			} else {
				var base string
				if !typeField.Anonymous {
					base = typeField.Name + "."
				}
				for _, val := range newList {
					list = append(list, base+val)
				}
			}
		}

	case reflect.Slice:
		size := field.Len()

		for j := 1; j <= size; j++ {
			valueField := field.Index(j - 1)

			newList := keys(valueField.Interface())
			if newList == nil { // || len(newList) == 0 {
				list = append(list, strconv.Itoa(j))
			} else {
				for _, val := range newList {
					list = append(list, strconv.Itoa(j)+"."+val)
				}
			}
		}

	case reflect.String, reflect.Int, reflect.Bool:
		return nil

	default:
		return nil // Should we panic instead?

	}

	return list
}
