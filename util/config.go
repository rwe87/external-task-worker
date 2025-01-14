/*
 * Copyright 2019 InfAI (CC SES)
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *    http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package util

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"reflect"
	"regexp"
	"strconv"
	"strings"
)

type ConfigStruct struct {
	DeviceRepoUrl			 string
	CamundaWorkerTimeout     int64
	CamundaWorkerTasks       int64
	CamundaFetchLockDuration int64
	CamundaUrl               string
	CamundaTopic             string
	ZookeeperUrl             string //host1:2181,host2:2181/chroot
	KafkaConsumerGroup       string
	ResponseTopic            string
	QosStrategy              string // <=, >=
	KafkaTimeout             int64
	SaramaLog                string
	FatalKafkaErrors         string
	AuthExpirationTimeBuffer float64
	AuthEndpoint             string
	AuthClientId             string
	AuthClientSecret         string
	JwtPrivateKey            string
	JwtExpiration            int64
	JwtIssuer                string
	PermissionsUrl           string
}
type ConfigType *ConfigStruct

var Config ConfigType

func LoadConfig(location string) error {
	file, error := os.Open(location)
	if error != nil {
		log.Println("error on config load: ", error)
		return error
	}
	decoder := json.NewDecoder(file)
	configuration := ConfigStruct{}
	error = decoder.Decode(&configuration)
	if error != nil {
		log.Println("invalid config json: ", error)
		return error
	}
	HandleEnvironmentVars(&configuration)
	HandleDefaultValues(&configuration)
	Config = &configuration
	return nil
}

func HandleDefaultValues(config ConfigType) {

}

var camel = regexp.MustCompile("(^[^A-Z]*|[A-Z]*)([A-Z][^A-Z]+|$)")

func fieldNameToEnvName(s string) string {
	var a []string
	for _, sub := range camel.FindAllStringSubmatch(s, -1) {
		if sub[1] != "" {
			a = append(a, sub[1])
		}
		if sub[2] != "" {
			a = append(a, sub[2])
		}
	}
	return strings.ToUpper(strings.Join(a, "_"))
}

// preparations for docker
func HandleEnvironmentVars(config ConfigType) {
	configValue := reflect.Indirect(reflect.ValueOf(config))
	configType := configValue.Type()
	for index := 0; index < configType.NumField(); index++ {
		fieldName := configType.Field(index).Name
		envName := fieldNameToEnvName(fieldName)
		envValue := os.Getenv(envName)
		if envValue != "" {
			fmt.Println("use environment variable: ", envName, " = ", envValue)
			if configValue.FieldByName(fieldName).Kind() == reflect.Int64 {
				i, _ := strconv.ParseInt(envValue, 10, 64)
				configValue.FieldByName(fieldName).SetInt(i)
			}
			if configValue.FieldByName(fieldName).Kind() == reflect.Float64 {
				f, _ := strconv.ParseFloat(envValue, 64)
				configValue.FieldByName(fieldName).SetFloat(f)
			}
			if configValue.FieldByName(fieldName).Kind() == reflect.String {
				configValue.FieldByName(fieldName).SetString(envValue)
			}
			if configValue.FieldByName(fieldName).Kind() == reflect.Slice {
				val := []string{}
				for _, element := range strings.Split(envValue, ",") {
					val = append(val, strings.TrimSpace(element))
				}
				configValue.FieldByName(fieldName).Set(reflect.ValueOf(val))
			}
			if configValue.FieldByName(fieldName).Kind() == reflect.Map {
				value := map[string]string{}
				for _, element := range strings.Split(envValue, ",") {
					keyVal := strings.Split(element, ":")
					key := strings.TrimSpace(keyVal[0])
					val := strings.TrimSpace(keyVal[1])
					value[key] = val
				}
				configValue.FieldByName(fieldName).Set(reflect.ValueOf(value))
			}
		}
	}
}
