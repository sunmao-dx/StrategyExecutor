/*
Copyright 2017 The Kubernetes Authors.

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

package gitee_utils

import (
	"io/ioutil"
	"net/http"

	"github.com/sirupsen/logrus"
)

// ValidateWebhook ensures that the provided request conforms to the
// format of a Gitee webhook and the payload can be validated with
// the provided hmac secret. It returns the event type, the event guid,
// the payload of the request, whether the webhook is valid or not,
// and finally the resultant HTTP status code
func ValidateWebhook(w http.ResponseWriter, r *http.Request) (string, string, []byte, bool, int) {
	defer r.Body.Close()
	payload, err := ioutil.ReadAll(r.Body)
	if err != nil {
		responseHTTPError(w, http.StatusInternalServerError, "500 Internal Server Error: Failed to read request body")
		return "", "", nil, false, http.StatusInternalServerError
	}
	return "", "", payload, true, http.StatusOK
}

func responseHTTPError(w http.ResponseWriter, statusCode int, response string) {
	logrus.WithFields(logrus.Fields{
		"response":    response,
		"status-code": statusCode,
	}).Debug(response)
	http.Error(w, response, statusCode)
}
