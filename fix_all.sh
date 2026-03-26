cd server/pkg/app
sed -i 's/m.Called(/m.Mock.Called(/g' api_audit_test.go
sed -i 's/mockStore.On(/mockStore.Mock.On(/g' api_audit_test.go
sed -i 's/mockRegistry.On(/mockRegistry.Mock.On(/g' api_handlers_extra_test.go
sed -i 's/m.Called(/m.Mock.Called(/g' api_test.go
sed -i 's/mockRegistry.On(/mockRegistry.Mock.On(/g' api_test.go
sed -i 's/mockRegistry.On(/mockRegistry.Mock.On(/g' dashboard_metrics_test.go
sed -i 's/m.Called(/m.Mock.Called(/g' dashboard_stats_integration_test.go
sed -i 's/mockRegistry.On(/mockRegistry.Mock.On(/g' dashboard_stats_integration_test.go
sed -i 's/m.Called(/m.Mock.Called(/g' server_init_test.go
sed -i 's/mockStore.On(/mockStore.Mock.On(/g' server_init_test.go
sed -i 's/mockStore.AssertExpectations(/mockStore.Mock.AssertExpectations(/g' server_init_test.go
sed -i 's/mockStore.AssertNotCalled(/mockStore.Mock.AssertNotCalled(/g' server_init_test.go
