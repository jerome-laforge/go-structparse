package parse

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"testing"

	"gopkg.in/yaml.v2"

	"github.com/stretchr/testify/assert"
)

type Embedded struct {
	KeyThree string
}

type FakeConfig struct {
	KeyOne   string
	Map      map[string]string
	Embedded Embedded
}

type HostExample struct {
	Name string
	Port string
}

type ConfigExample struct {
	Name string
	Host HostExample
}

const (
	keyOneName    = "KEY_ONE"
	keyOneValue   = "KEY_ONE_VALUE"
	keyTwoIndex   = "KEY_TWO_INDEX"
	keyTwoName    = "KEY_TWO"
	keyTwoValue   = "KEY_TWO_VALUE"
	keyThreeName  = "KEY_THREE"
	keyThreeValue = "KEY_THREE_VALUE"
)

const (
	prefix    = "TEST_"
	delimiter = "$"
)

func GetKey(i uint64) string {
	return fmt.Sprintf("KEY%d", i)
}

func GetValue(i uint64) string {
	return fmt.Sprintf("VALUE%d", i)
}

var keys = make(map[string]string)

func RegisterKey(i uint64) {
	name := fmt.Sprintf("%s%s", prefix, GetKey(i))
	keys[name] = GetValue(i)
}

func ClearKey(i uint64) {
	name := fmt.Sprintf("%s%s", prefix, GetKey(i))
	keys[name] = ""
}

func FakeEnvMapper(line string) string {
	key := fmt.Sprintf("%s%s", prefix, strings.Replace(line, delimiter, "", -1))
	val, ok := keys[key]
	if !ok {
		return "ERROR"
	}
	return val
}

func NoTestConfigInfill(t *testing.T) {

	t.Run("Handles struct fields", func(t *testing.T) {

		RegisterKey(1)

		c := struct {
			Test string
		}{
			GetKey(1),
		}

		StructStrings(FakeEnvMapper, &c)

		assert.EqualValues(t, GetValue(1), c.Test)
	})

	t.Run("Infills strings in structure from environment", func(t *testing.T) {
		prefix := "TEST_"
		delimiter := "$"

		c := FakeConfig{
			KeyOne:   fmt.Sprintf("%s%s", delimiter, keyOneName),
			Map:      make(map[string]string),
			Embedded: Embedded{fmt.Sprintf("%s%s", delimiter, keyThreeName)},
		}
		c.Map[keyTwoIndex] = fmt.Sprintf("%s%s", delimiter, keyTwoName)

		os.Setenv(fmt.Sprintf("%s%s", prefix, keyOneName), keyOneValue)
		os.Setenv(fmt.Sprintf("%s%s", prefix, keyTwoName), keyTwoValue)
		os.Setenv(fmt.Sprintf("%s%s", prefix, keyThreeName), keyThreeValue)

		StructStrings(NewEnvironmentMapper(delimiter, prefix), &c)

		assert.EqualValues(t, keyOneValue, c.KeyOne)
		assert.EqualValues(t, keyTwoValue, c.Map[keyTwoIndex])
		assert.EqualValues(t, keyThreeValue, c.Embedded.KeyThree)
	})

	t.Run("Loads and infill structures from file", func(t *testing.T) {
		delimiter := "$"
		prefix := "TEST_"

		os.Setenv(fmt.Sprintf("%s%s", prefix, "NAME"), "APP_NAME")
		os.Setenv(fmt.Sprintf("%s%s", prefix, "HOSTNAME"), "localhost")
		os.Setenv(fmt.Sprintf("%s%s", prefix, "PORT"), "9009")

		c := ConfigExample{}

		// Load configuration file
		data, err := ioutil.ReadFile("./example.yml")
		assert.Nil(t, err)

		// Unmarshal from yaml
		err = yaml.Unmarshal(data, &c)
		assert.Nil(t, err)

		StructStrings(NewEnvironmentMapper(delimiter, prefix), &c)

		assert.EqualValues(t, "APP_NAME", c.Name)
		assert.EqualValues(t, "localhost", c.Host.Name)
		assert.EqualValues(t, "9009", c.Host.Port)
	})

}
