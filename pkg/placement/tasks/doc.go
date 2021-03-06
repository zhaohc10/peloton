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

package tasks

/*
Package tasks contains the task service interface and an implementation of it. The task service is responsible for
dequeueing gangs/tasks for placement, setting placements of tasks on offers in the resource manager and for enqueueing
gangs/tasks, which failed to get placed, back into the resource manager. The task service is also responsible for any
kind of logging and metrics emission so that these things will not pollute the code in the placement engine main loop.
*/
