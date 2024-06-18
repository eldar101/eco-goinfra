package secret

import (
	"testing"
        "fmt"
	"github.com/openshift-kni/eco-goinfra/pkg/clients"
	operatorv1 "github.com/openshift/api/operator/v1"
	corev1 "k8s.io/api/core/v1"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

var (
	defaultSecretName      = "test-name"
	defaultSecretNamespace = "test-namespace"
	defaultSecretType = "test-secrettype"
)

func TestSecretPull(t *testing.T) {
	testCases := []struct {
		secretName         string
		secretNamespace    string
		expectedError       bool
		addToRuntimeObjects bool
		client              bool
		expectedErrorText   string
	}{
		{
                        secretName:          defaultSecretName,
                        secretNamespace:     defaultSecretNamespace,
			expectedError:       false,
			addToRuntimeObjects: true,
			client:              true,
			expectedErrorText:   "",
		},
		{
                        secretName:          defaultSecretName,
                        secretNamespace:     defaultSecretNamespace,
			expectedError:       true,
			addToRuntimeObjects: false,
			client:              true,
			expectedErrorText:   "secret object test not found in namespace test",

		},
		{
                        secretName:          "",
                        secretNamespace:     defaultSecretNamespace,
			expectedError:       true,
			addToRuntimeObjects: true,
			client:              true,
			expectedErrorText:   "secret name cannot be empty",
		},
		{
                        secretName:          defaultSecretName,
                        secretNamespace:     "",
			expectedError:       true,
			addToRuntimeObjects: true,
			client:              true,
			expectedErrorText:   "secret namespace cannot be empty",
		},
                {
                        secretName:          defaultSecretName,
                        secretNamespace:     defaultSecretNamespace,
			expectedError:       true,
                        addToRuntimeObjects: true,
			client:              false,
                        expectedErrorText:   "secret client cannot be empty",

                },
	}

	for _, testCase := range testCases {
		var runtimeObjects []runtime.Object

		var testSettings *clients.Settings

		if testCase.addToRuntimeObjects {
			runtimeObjects = append(runtimeObjects, &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      testCase.secretName,
					Namespace: testCase.secretNamespace,
				},
			})
		}

		testSettings = clients.GetTestClients(clients.TestClientParams{
			K8sMockObjects: runtimeObjects,
		})

                if testCase.client {
                        testSettings = clients.GetTestClients(clients.TestClientParams{
                                K8sMockObjects: runtimeObjects,
                        })
                }

		builderResult, err := Pull(testSettings, testCase.secretName, testCase.secretNamespace)

		if testCase.expectedError {
			assert.NotNil(t, err)

			if testCase.expectedErrorText != "" {
				assert.Equal(t, testCase.expectedErrorText, err.Error())
			}
		} else {
			assert.Nil(t, err)
			assert.NotNil(t, builderResult)
		}
	}
}

func TestSecretNewBuilder(t *testing.T) {
        testCases := []struct {
                name          string
                namespace     string
                secretType    string
                expectedError string
        }{
                {
                        name:          defaultSecretName,
                        namespace:     defaultSecretNamespace,
                        secretType:    defaultSecretType,
                        expectedError: "",
                },
                {
                        name:          "",
                        namespace:     defaultSecretNamespace,
                        secretType:    defaultSecretType,
                        expectedError: "secret 'name' cannot be empty",
                },
                {
                        name:          defaultSecretName,
                        namespace:     "",
                        secretType:    defaultSecretType,
                        expectedError: "secret 'nsname' cannot be empty",
                },
                {
                        name:          defaultSecretName,
                        namespace:     defaultSecretNamespace,
                        secretType:    "",
                        expectedError: "secret 'secretType' cannot be empty",
                },

        }

        for _, testCase := range testCases {
                testSettings := clients.GetTestClients(clients.TestClientParams{})
                testSecretBuilder := NewBuilder(testSettings, testCase.name, testCase.namespace, testCase.secretType)
                assert.Equal(t, testCase.expectedError, testSecretBuilder.errorMsg)
                assert.NotNil(t, testSecretBuilder.Definition)

                if testCase.expectedError == "" {
                        assert.Equal(t, testCase.name, testSecretBuilder.Definition.Name)
                        assert.Equal(t, testCase.namespace, testSecretBuilder.Definition.Namespace)
                }
        }
}

func TestSecretCreate(t *testing.T) {
        testCases := []struct {
                testSecret *SecretBuilder
                expectedError     error
        }{
                {
                        testSecret: buildValidSecretBuilder(buildSecretWithDummyObject()),
                        expectedError:     nil,
                },
                {
                        testSecret: buildInValidSecretBuilder(buildTestClientWithDummyObject()),
                        expectedError:     fmt.Errorf("Secret 'test-secret' cannot be empty list"),
                },
        }

        for _, testCase := range testCases {
                SecretBuilder, err := testCase.testSecret.Create()
                assert.Equal(t, err, testCase.expectedError)

                if testCase.expectedError == nil {
                        assert.Equal(t, SecretBuilder.Definition, SecretBuilder.Object)
                }
        }
}

func TestSecretDelete(t *testing.T) {
	testCases := []struct {
		secretExistsAlready bool
		name                 string
		namespace            string
	}{
		{
			secretExistsAlready: true,
			name:                 defaultSecretName,
			namespace:            defaultSecretNamespace,
		},
		{
			secretExistsAlready: false,
			name:                 defaultSecretName,
			namespace:            defaultSecretNamespace,
		},
	}

	for _, testCase := range testCases {
		var runtimeObjects []runtime.Object

		if testCase.secretExistsAlready {
			runtimeObjects = append(runtimeObjects, &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      testCase.name,
					Namespace: testCase.namespace,
				},
			})
		}

		testBuilder, client := buildTestBuilderWithFakeObjects(runtimeObjects, testCase.name, testCase.namespace)

		err := testBuilder.Delete()
		assert.Nil(t, err)

		// Assert that the object actually does not exist
		_, err = Pull(client, testCase.name, testCase.namespace)
		assert.NotNil(t, err)
	}
}

func TestSecretExists(t *testing.T) {
	testCases := []struct {
		secretExistsAlready bool
		name                 string
		namespace            string
	}{
		{
			secretExistsAlready: true,
			name:                 defaultSecretName,
			namespace:            defaultSecretNamespace,
		},
		{
			secretExistsAlready: false,
			name:                 defaultSecretName,
			namespace:            defaultSecretNamespace,
		},
	}

	for _, testCase := range testCases {
		var runtimeObjects []runtime.Object

		if testCase.secretExistsAlready {
			runtimeObjects = append(runtimeObjects, &operatorv1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      testCase.name,
					Namespace: testCase.namespace,
				},
			})
		}

		testBuilder, _ := buildTestBuilderWithFakeObjects(runtimeObjects, testCase.name, testCase.namespace)

		result := testBuilder.Exists()
		if testCase.secretExistsAlready {
			assert.True(t, result)
		} else {
			assert.False(t, result)
		}
	}
}
func TestSecretWithOptions(t *testing.T) {
	testSettings := buildSecretWithDummyObject()
	testBuilder := buildValidSecretHostBuilder(testSettings).WithOptions(
		func(builder *SecretBuilder) (*BmhBuilder, error) {
			return builder, nil
		})

	assert.Equal(t, "", testBuilder.errorMsg)
	testBuilder = buildValidSecretBuilder(testSettings).WithOptions(
		func(builder *SecretBuilder) (*SecretBuilder, error) {
			return builder, fmt.Errorf("error")
		})
	assert.Equal(t, "error", testBuilder.errorMsg)
}

func TestSecretWithData(t *testing.T) {
	testCases := []struct {
		key         string
		value       string
		expectedErr string
	}{
		{
			key:         "key",
			value:       "value",
			expectedErr: "",
		},
		{
			key:         "",
			value:       "",
			expectedErr: "'data' cannot be empty",
		},
	}

	for _, testCase := range testCases {
		testBuilder := buildTestBuilderWithFakeObjects([]runtime.Object{})

		if testCase.expectedErr == "" {
			testBuilder.WithData(map[string]string{testCase.key: testCase.value})

			assert.Equal(t, testCase.value, testBuilder.Definition.Data[testCase.key])
		} else {
			testBuilder.WithData(map[string]string{})

			assert.Equal(t, testCase.expectedErr, testBuilder.errorMsg)
		}
	}
}

func TestSecretWithAnnotations(t *testing.T) {
	testCases := []struct {
		testAnnotations   map[string]string
		expectedErrorText string
	}{
		{
			testAnnotations:   map[string]string{"openshift.io/internal-registry-auth-token.binding": "bound"},
			expectedErrorText: "",
		},
		{
			testAnnotations:   map[string]string{"openshift.io/internal-registry-auth-token.service-account": "default"},
			expectedErrorText: "",
		},
		{
			testAnnotations:   map[string]string{},
			expectedErrorText: "'annotations' argument cannot be empty",
		},
	}

	for _, testCase := range testCases {
		testBuilder := buildValidSecretBuilder(buildSecretWithDummyObject())

		testBuilder.WithAnnotation(testCase.testAnnotation)

		assert.Equal(t, testCase.expectedErrMsg, testBuilder.errorMsg)

		if testCase.expectedErrMsg == "" {
			assert.Equal(t, testCase.testAnnotation, testBuilder.Definition.Annotations)
		}
	}
}

func TestSecretUpdate(t *testing.T) {
	generateTestSecret := func() *appsv1.Secret {
		return &appsv1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      defaultSecretName,
				Namespace: defaultSecretNamespace,
			},
		}
	}

	testCases := []struct {
		secretExistsAlready bool
	}{
		{
			secretExistsAlready: false,
		},
		{
			secretExistsAlready: true,
		},
	}

	for _, testCase := range testCases {
		var runtimeObjects []runtime.Object

		if testCase.secretExistsAlready {
			runtimeObjects = append(runtimeObjects, generateTestSecret())
		}

		testBuilder := buildTestBuilderWithFakeObjects(runtimeObjects)

		// Assert the secret before the update
		assert.NotNil(t, testBuilder.Definition)
		assert.Equal(t, "value", string(testBuilder.Definition.Data["key"]))

		// Set a value in the definition to test the update
		testBuilder.Definition.Data["key"] = []byte("test")

		// Perform the update
		result, err := testBuilder.Update()

		// Assert the result
		assert.NotNil(t, testBuilder.Definition)

		if !testCase.secretExistsAlready {
			assert.NotNil(t, err)
			assert.Nil(t, result.Object)
		} else {
			assert.Nil(t, err)
			assert.Equal(t, testBuilder.Definition.Name, result.Definition.Name)
			assert.Equal(t, "test", string(result.Definition.Data["key"]))
		}
	}
}

func buildTestBuilderWithFakeObjects(runtimeObjects []runtime.Object,
	name, namespace string) (*Builder, *clients.Settings) {
	testSettings := clients.GetTestClients(clients.TestClientParams{
		K8sMockObjects: runtimeObjects,
	})

	return &Builder{
		apiClient: testSettings,
		Definition: &operatorv1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      name,
				Namespace: namespace,
				Type:      secretType,
			},
		},
	}, testSettings
}

func buildSecretBuilder(apiClient *clients.Settings) *Builder {
	secretBuilder := NewBuilder(
		apiClient,
		defaultSecretName,
		defaultSecretNamespace,
		defaultSecretType)

	return secretBuilder
}

func buildInValidSecretBuilder(apiClient *clients.Settings) *Builder {
	secretBuilder := &Builder{
		apiClient: apiClient,
		Definition: &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      defaultSecretName,
				Namespace: defaultSecretNamespace,         
			},
		},
	}
	return secretBuilder
}


func buildTestClientWithDummyObject() *clients.Settings {
        return clients.GetTestClients(clients.TestClientParams{
                K8sMockObjects: buildDummySecret(),
        })
}

func buildDummySecret() []runtime.Object {
        return append([]runtime.Object{}, &mlbtypes.Secret{
                ObjectMeta: metav1.ObjectMeta{
                        Name:      defaultSecretName,
                        Namespace: defaultSecretNamespace,
                },
        })
}

func TestSecretValidate(t *testing.T) {
        testCases := []struct {
                builderNil    bool
                definitionNil bool
                apiClientNil  bool
                expectedError string
        }{
                {
                        builderNil:    true,
                        definitionNil: false,
                        apiClientNil:  false,
                        expectedError: "error: received nil Secret builder",
                },
                {
                        builderNil:    false,
                        definitionNil: true,
                        apiClientNil:  false,
                        expectedError: "can not redefine the undefined Secret",
                },
                {
                        builderNil:    false,
                        definitionNil: false,
                        apiClientNil:  true,
                        expectedError: "Secret builder cannot have nil apiClient",
                },
                {
                        builderNil:    false,
                        definitionNil: false,
                        apiClientNil:  false,
                        expectedError: "",
                },
        }

        for _, testCase := range testCases {
                testBuilder, _ := buildTestBuilderWithFakeObjects(nil, "test", "test")

                if testCase.builderNil {
                        testBuilder = nil
                }

                if testCase.definitionNil {
                        testBuilder.Definition = nil
                }

                if testCase.apiClientNil {
                        testBuilder.apiClient = nil
                }

                result, err := testBuilder.validate()
                if testCase.expectedError != "" {
                        assert.NotNil(t, err)
                        assert.Equal(t, testCase.expectedError, err.Error())
                        assert.False(t, result)
                } else {
                        assert.Nil(t, err)
                        assert.True(t, result)
                }
        }
}
