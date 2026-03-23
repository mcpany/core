sed -i 's/mockStore.On("Read", mock.Anything, mock.Anything), mock.Anything, mock.Anything).Return(entries, nil).Once()/mockStore.On("Read", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(entries, nil).Once()/g' server/pkg/app/api_audit_test.go
sed -i 's/mockStore.AssertExpectations(t)(t)/mockStore.AssertExpectations(t)/g' server/pkg/app/server_init_test.go
