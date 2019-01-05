/*

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package common

import (
	"bytes"
	"text/template"
)

// ControllerLabels returns common labels to apply
var ControllerLabels = map[string]string{
	"app.kubernetes.io/managed-by": "drupal-operator.sylus.ca",
}

// GenerateConfig returns rendered template against given interface
func GenerateConfig(TemplateInput interface{}, configmapTemplate string) *string {
	output := new(bytes.Buffer)
	tpl, err := template.New("config").Delims("[[", "]]").Parse(configmapTemplate)
	if err != nil {
		return nil
	}
	err = tpl.Execute(output, TemplateInput)
	if err != nil {
		return nil
	}
	outputString := output.String()
	return &outputString
}
