// Copyright (c) 2019 Uber Technologies, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package goalstate

import (
	"testing"

	"github.com/uber/peloton/.gen/peloton/api/v0/job"

	"github.com/stretchr/testify/assert"
)

func TestGetDefaultJobGoalState(t *testing.T) {
	assert.Equal(t,
		GetDefaultJobGoalState(job.JobType_SERVICE),
		job.JobState_RUNNING)

	assert.Equal(t,
		GetDefaultJobGoalState(job.JobType_BATCH),
		job.JobState_SUCCEEDED)
}
