package cassandra

import (
	"context"
	"testing"

	"code.uber.internal/infra/peloton/.gen/peloton/api/v0/job"
	"code.uber.internal/infra/peloton/.gen/peloton/api/v0/peloton"
	"code.uber.internal/infra/peloton/.gen/peloton/api/v0/task"
	"code.uber.internal/infra/peloton/storage"
	datastore "code.uber.internal/infra/peloton/storage/cassandra/api"
	datastoremocks "code.uber.internal/infra/peloton/storage/cassandra/api/mocks"
	datastoreimpl "code.uber.internal/infra/peloton/storage/cassandra/impl"

	"github.com/golang/mock/gomock"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/suite"
)

const (
	testJob = "941ff353-ba82-49fe-8f80-fb5bc649b04d"
)

type MockDatastoreTestSuite struct {
	suite.Suite
	testJobID *peloton.JobID

	ctrl            *gomock.Controller
	mockedDataStore *datastoremocks.MockDataStore
	store           *Store
}

func (suite *MockDatastoreTestSuite) SetupTest() {
	var result datastore.ResultSet

	suite.testJobID = &peloton.JobID{Value: testJob}

	suite.ctrl = gomock.NewController(suite.T())
	suite.mockedDataStore = datastoremocks.NewMockDataStore(suite.ctrl)
	conf := MigrateForTest()

	suite.store = &Store{
		DataStore:   suite.mockedDataStore,
		metrics:     storage.NewMetrics(testScope.SubScope("storage")),
		Conf:        conf,
		retryPolicy: nil,
	}

	queryBuilder := &datastoreimpl.QueryBuilder{}
	// Mock datastore execute to fail
	suite.mockedDataStore.EXPECT().Execute(gomock.Any(), gomock.Any()).
		Return(result, errors.New("my-error")).AnyTimes()
	suite.mockedDataStore.EXPECT().NewQuery().Return(queryBuilder).AnyTimes()
}

func TestMockDatastoreTestSuite(t *testing.T) {
	suite.Run(t, new(MockDatastoreTestSuite))
}

// TestDataStoreFailureGetJobConfig tests datastore failures in getting job cfg
func (suite *MockDatastoreTestSuite) TestDataStoreFailureGetJobConfig() {
	_, err := suite.store.GetJobConfigWithVersion(
		context.Background(), suite.testJobID, 0)
	suite.Error(err)

	_, err = suite.store.GetJobConfig(
		context.Background(), suite.testJobID)
	suite.Error(err)

}

// TestDataStoreFailureGetJobRuntime tests datastore failures in getting
// job runtime
func (suite *MockDatastoreTestSuite) TestDataStoreFailureGetJobRuntime() {
	_, err := suite.store.GetJobRuntime(
		context.Background(), suite.testJobID)
	suite.Error(err)
}

// TestDataStoreFailureGetJobSummary tests datastore failures in getting
// job summary
func (suite *MockDatastoreTestSuite) TestDataStoreFailureGetJobSummary() {
	_, err := suite.store.getJobSummaryFromIndex(
		context.Background(), suite.testJobID)
	suite.Error(err)

	_, err = suite.store.getJobSummaryFromConfig(
		context.Background(), suite.testJobID)
	suite.Error(err)
}

// TestDataStoreFailureGetJob tests datastore failures in getting job
func (suite *MockDatastoreTestSuite) TestDataStoreFailureGetJob() {
	_, err := suite.store.GetJobsByStates(
		context.Background(), []job.JobState{job.JobState_RUNNING})
	suite.Error(err)

	_, err = suite.store.GetMaxJobConfigVersion(
		context.Background(), suite.testJobID)
	suite.Error(err)
}

// TestDataStoreFailureTaskEvents tests datastore failures in getting task evt
func (suite *MockDatastoreTestSuite) TestDataStoreFailureTaskEvents() {
	_, err := suite.store.GetTaskStateChanges(
		context.Background(), suite.testJobID, 0)
	suite.Error(err)
	_, err = suite.store.GetTaskEvents(
		context.Background(), suite.testJobID, 0)
	suite.Error(err)
}

// TestDataStoreFailureGetTasks tests datastore failures in getting tasks
func (suite *MockDatastoreTestSuite) TestDataStoreFailureGetTasks() {
	_, err := suite.store.GetTasksForJobAndStates(
		context.Background(), suite.testJobID, []task.TaskState{
			task.TaskState(task.TaskState_PENDING)})
	suite.Error(err)

	_, err = suite.store.GetTasksForJobResultSet(
		context.Background(), suite.testJobID)
	suite.Error(err)

	_, err = suite.store.GetTasksForJob(
		context.Background(), suite.testJobID)
	suite.Error(err)

	_, err = suite.store.GetTaskForJob(
		context.Background(), suite.testJobID, 0)
	suite.Error(err)

	_, err = suite.store.GetTaskIDsForJobAndState(
		context.Background(), suite.testJobID, task.TaskState_PENDING.String())
	suite.Error(err)

	_, err = suite.store.getTaskStateCount(
		context.Background(), suite.testJobID, task.TaskState_PENDING.String())
	suite.Error(err)

	_, err = suite.store.getTask(context.Background(), testJob, 0)
	suite.Error(err)
}

// TestDataStoreFailureGetTaskConfig tests datastore failures in getting task cfg
func (suite *MockDatastoreTestSuite) TestDataStoreFailureGetTaskConfig() {
	_, err := suite.store.GetTaskConfig(
		context.Background(), suite.testJobID, 0, 0)
	suite.Error(err)

	_, err = suite.store.GetTaskConfigs(
		context.Background(), suite.testJobID, []uint32{0}, 0)
	suite.Error(err)

	_, err = suite.store.GetTaskStateSummaryForJob(
		context.Background(), suite.testJobID)
	suite.Error(err)
}

// TestDataStoreFailureGetTaskRuntime tests datastore failures in getting
// task runtime
func (suite *MockDatastoreTestSuite) TestDataStoreFailureGetTaskRuntime() {
	_, err := suite.store.GetTaskRuntimesForJobByRange(
		context.Background(), suite.testJobID, &task.InstanceRange{
			From: uint32(0),
			To:   uint32(3),
		})
	suite.Error(err)

	_, err = suite.store.GetTaskRuntime(
		context.Background(), suite.testJobID, 0)
	suite.Error(err)

	_, err = suite.store.getTaskRuntimeRecord(context.Background(), testJob, 0)
	suite.Error(err)
}

// TestDataStoreFailureJobQuery tests datastore failures in job query
func (suite *MockDatastoreTestSuite) TestDataStoreFailureJobQuery() {
	_, _, _, err := suite.store.QueryJobs(
		context.Background(), nil, &job.QuerySpec{}, false)
	suite.Error(err)
}

// TestDataStoreFailureTaskQuery tests datastore failures in task query
func (suite *MockDatastoreTestSuite) TestDataStoreFailureTaskQuery() {
	_, _, err := suite.store.QueryTasks(
		context.Background(), suite.testJobID, &task.QuerySpec{})
	suite.Error(err)
}

// TestDataStoreFailureGetRespools tests datastore failures in get respools
func (suite *MockDatastoreTestSuite) TestDataStoreFailureGetRespools() {
	_, err := suite.store.GetAllResourcePools(context.Background())
	suite.Error(err)

	_, err = suite.store.GetResourcePoolsByOwner(context.Background(), "dummy")
	suite.Error(err)
}

// TestDataStoreFailureFramework tests datastore failures in get frameworks
func (suite *MockDatastoreTestSuite) TestDataStoreFailureFramework() {
	_, err := suite.store.getFrameworkInfo(context.Background(), "framwork-id")
	suite.Error(err)
}

// TestDataStoreFailureGetPersistentVolume tests datastore failures in get
// persistent volume
func (suite *MockDatastoreTestSuite) TestDataStoreFailureGetPersistentVolume() {
	_, err := suite.store.GetPersistentVolume(
		context.Background(), &peloton.VolumeID{Value: "test"})
	suite.Error(err)
}

// TestDataStoreFailureGetSecret tests datastore failures in get secret
func (suite *MockDatastoreTestSuite) TestDataStoreFailureGetSecret() {
	_, err := suite.store.GetSecret(
		context.Background(), &peloton.SecretID{Value: "test"})
	suite.Error(err)
}

// TestDataStoreFailureGetUpdate tests datastore failures in get update
func (suite *MockDatastoreTestSuite) TestDataStoreFailureGetUpdate() {
	_, err := suite.store.GetUpdate(
		context.Background(), &peloton.UpdateID{Value: "test"})
	suite.Error(err)

	_, err = suite.store.GetUpdateProgress(
		context.Background(), &peloton.UpdateID{Value: "test"})
	suite.Error(err)

	_, err = suite.store.GetUpdatesForJob(context.Background(), suite.testJobID)
	suite.Error(err)
}

// TestDataStoreFailureDeleteJobCfgVersion tests datastore failures in delete
// job config version
func (suite *MockDatastoreTestSuite) TestDataStoreFailureDeleteJobCfgVersion() {
	ctx := context.Background()
	var result datastore.ResultSet

	// Setup mocks for this context

	// Simulate failure to delete task config
	suite.mockedDataStore.EXPECT().Execute(ctx, gomock.Any()).
		Return(result, errors.New("my-error"))

	err := suite.store.deleteJobConfigVersion(ctx, suite.testJobID, 0)
	suite.Error(err)

	// Simulate success to to delete task cfg and failure to delete job cfg
	suite.mockedDataStore.EXPECT().Execute(ctx, gomock.Any()).
		Return(result, nil)
	suite.mockedDataStore.EXPECT().Execute(ctx, gomock.Any()).
		Return(result, errors.New("my-error"))

	err = suite.store.deleteJobConfigVersion(ctx, suite.testJobID, 0)
	suite.Error(err)
}
