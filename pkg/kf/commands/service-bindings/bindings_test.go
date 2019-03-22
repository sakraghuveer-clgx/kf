package servicebindings_test

import (
	"errors"
	"testing"

	servicebindingscmd "github.com/GoogleCloudPlatform/kf/pkg/kf/commands/service-bindings"
	"github.com/GoogleCloudPlatform/kf/pkg/kf/internal/testutil"
	servicebindings "github.com/GoogleCloudPlatform/kf/pkg/kf/service-bindings"
	"github.com/GoogleCloudPlatform/kf/pkg/kf/service-bindings/fake"
	"github.com/golang/mock/gomock"
	"github.com/poy/service-catalog/pkg/apis/servicecatalog/v1beta1"
)

func TestNewListBindingsCommand(t *testing.T) {
	cases := map[string]serviceTest{
		"wrong number of args": {
			Args:        []string{"FOO"},
			ExpectedErr: errors.New("accepts 0 arg(s), received 1"),
		},
		"command params get passed correctly": {
			Args:      []string{"--app=APP_NAME", "--service=SERVICE_INSTANCE"},
			Namespace: "custom-ns",
			Setup: func(t *testing.T, f *fake.FakeClientInterface) {
				f.EXPECT().List(gomock.Any()).Do(func(opts ...servicebindings.ListOption) {
					config := servicebindings.ListOptions(opts)
					testutil.AssertEqual(t, "namespace", "custom-ns", config.Namespace())
					testutil.AssertEqual(t, "app name", "custom-ns", config.Namespace())
					testutil.AssertEqual(t, "service instance name", "custom-ns", config.Namespace())
				}).Return([]v1beta1.ServiceBinding{}, nil)
			},
		},
		"defaults config": {
			Args: []string{},
			Setup: func(t *testing.T, f *fake.FakeClientInterface) {
				f.EXPECT().List(gomock.Any()).Do(func(opts ...servicebindings.ListOption) {
					config := servicebindings.ListOptions(opts)
					testutil.AssertEqual(t, "namespace", "", config.Namespace())
				}).Return([]v1beta1.ServiceBinding{}, nil)
			},
		},
		"bad server call": {
			Args: []string{},
			Setup: func(t *testing.T, f *fake.FakeClientInterface) {
				f.EXPECT().List(gomock.Any()).Return(nil, errors.New("api-error"))
			},
			ExpectedErr: errors.New("api-error"),
		},
		"output list contains items": {
			Args: []string{},
			Setup: func(t *testing.T, f *fake.FakeClientInterface) {
				f.EXPECT().List(gomock.Any()).Return([]v1beta1.ServiceBinding{
					*dummyBindingInstance("app1", "instance1"),
					*dummyBindingInstance("app2", "instance2"),
				}, nil)
			},
			ExpectedStrings: []string{"app1", "instance1", "app2", "instance2"},
		},
	}

	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {
			runTest(t, tc, servicebindingscmd.NewListBindingsCommand)
		})
	}
}
