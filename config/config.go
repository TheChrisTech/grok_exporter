// Copyright 2016-2020 The grok_exporter Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package config

import (
	"fmt"
	"github.com/fstab/grok_exporter/config/v2"
	v3 "github.com/fstab/grok_exporter/config/v3"
	"io/ioutil"
	"regexp"
	"strconv"
	"strings"
	"github.com/jdrews/go-tailer/glob"
	"time"
)

// Example config: See ./example/config.yml

func LoadConfigFile(filename string) (*v3.Config, string, error) {
	content, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, "", fmt.Errorf("Failed to load %v: %v", filename, err.Error())
	}
	cfg, warn, err := LoadConfigString(content)
	if err != nil {
		return nil, warn, fmt.Errorf("Failed to load %v: %v", filename, err.Error())
	}
	return cfg, warn, nil
}

func LoadConfigString(content []byte) (*v3.Config, string, error) {
	version, warn, err := findVersion(string(content))
	if err != nil {
		return nil, warn, err
	}
	cfg, err := unmarshal(content, version)
	return cfg, warn, err
}

// returns (version, warning, error).
func findVersion(content string) (int, string, error) {
	warning := "Configuration version 2 found. This is still supported, but we recommend updating to version 3. Run grok_exporter with the -showconfig command line parameter to automatically convert to version 3 and write the result to the console."
	versionExpr := regexp.MustCompile(`"?global"?:\s*"?config_version"?:[\t\f ]*(\S+)`)
	versionInfo := versionExpr.FindStringSubmatch(content)
	if len(versionInfo) == 2 {
		version, err := strconv.Atoi(strings.TrimSpace(versionInfo[1]))
		if err != nil {
			return 0, "", fmt.Errorf("invalid 'global' configuration: '%v' is not a valid 'config_version'.", versionInfo[1])
		}
		if version == 2 {
			return version, warning, nil
		} else {
			return version, "", nil
		}
	} else { // no version found
		return 0, "", fmt.Errorf("invalid configuration: 'global.config_version' not found.")
	}
}

func unmarshal(content []byte, version int) (*v3.Config, error) {
	switch version {
	case 2:
		v2cfg, err := v2.Unmarshal(content)
		if err != nil {
			return nil, err
		}
		return v3.Convert(v2cfg)
	case 3:
		return v3.Unmarshal(content)
	default:
		return nil, fmt.Errorf("global.config_version %v is not supported", version)
	}
}

type InputConfig struct {
	Type                       string `yaml:",omitempty"`
	PathsAndGlobs              `yaml:",inline"`
	FailOnMissingLogfileString string        `yaml:"fail_on_missing_logfile,omitempty"` // cannot use bool directly, because yaml.v2 doesn't support true as default value.
	FailOnMissingLogfile       bool          `yaml:"-"`
	Readall                    bool          `yaml:",omitempty"`
	PollInterval               time.Duration `yaml:"poll_interval,omitempty"` // implicitly parsed with time.ParseDuration()
	MaxLinesInBuffer           int           `yaml:"max_lines_in_buffer,omitempty"`
	WebhookPath                string        `yaml:"webhook_path,omitempty"`
	WebhookFormat              string        `yaml:"webhook_format,omitempty"`
	WebhookJsonSelector        string        `yaml:"webhook_json_selector,omitempty"`
	WebhookTextBulkSeparator   string        `yaml:"webhook_text_bulk_separator,omitempty"`
	KafkaVersion               string        `yaml:"kafka_version,omitempty"`
	KafkaBrokers               []string      `yaml:"kafka_brokers,omitempty"`
	KafkaTopics                []string      `yaml:"kafka_topics,omitempty"`
	KafkaPartitionAssignor     string        `yaml:"kafka_partition_assignor,omitempty"`
	KafkaConsumerGroupName     string        `yaml:"kafka_consumer_group_name,omitempty"`
	KafkaConsumeFromOldest     bool          `yaml:"kafka_consume_from_oldest,omitempty"`
}

type PathsAndGlobs struct {
	Path  string      `yaml:",omitempty"`
	Paths []string    `yaml:",omitempty"`
	Globs []glob.Glob `yaml:"-"`
}
