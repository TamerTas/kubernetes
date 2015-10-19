# Generic Configuration Object

## Abstract

This proposal proposes a generic configuration object named ``ConfigData`` stored inside ``etcd`` that
stores data used for the configuration of ``Kubernetes`` components.

The main focus points of this proposal are:

* Solving dynamic configuration problem of components.
* Encapsulate configuration information to decrease component configuration complexity.
* Create a flexible configuration model for ``Kubernetes`` components.
* Increase overall system quality.

## Motivation

Currently command-line flags and environment variables are used for component configurations in ``Kubernetes``
which is complicated and error-prone. This approach is not suitable for a distributed computing environment.
Specifically the problems with the current approach are:

* Complicated deployment orchestration
* Hard to synchronize across replicated components.
* Hard to change dynamically as needed by continuous services.
* Hard to change in different versions that may support different configuration options.

## Use Cases

1. Be able to load API resources as environment variables into a component.
2. Be able to load API resources as mounted volumes into a component.
3. Be able to update mounted volumes after a change to the API resource.
4. Be able to configure components using an API resource instead of command-line flags.

## Solution

A new ``ConfigData`` API resource is proposed to address these problems.

The ``ConfigData`` architecture will be very similar to that of ``Secrets``.

``ConfigData`` can either be mounted as a volume or can be injected as environment
variables to a container through the use of the downward API.

### Dynamic Configuration

Any long-running system has mutating specification over time. In order to facilitate this functionality,
``ConfigData`` will be versioned and updates will automatically made available to the container (downward API).

``resourceVersion`` (Found in ``ObjectMetadata``) of the ``ConfigData`` object will be updated by the API server every time the object is modified.
After an update, modifications will be made visible to the consumer container. If the consumer uses the ``Data`` pairs
only for initialization or during starting process, A rolling-update might be necessary to update the components.

It is then the consumer component's responsibility to make use of the updated data
once it is made visible (or perform a rolling-update on consumers of that object).

### Advantages

* Reusable across different components.
* Easy distribution through the API server.
* Ability to use ``Kubernetes`` authentication and security policies with configuration objects.
* By leveraging the power of ``/watch`` API along with the new resource, we can change component configurations dynamically.
* Layer of abstraction that gets rid of the global state currently used for command-line flags and environment variables.

### API Resource

A new resource for ``ConfigData`` will be added to the API:

```go
	type ConfigData struct {
		TypeMeta   `json:",inline"`
		ObjectMeta `json:"metadata,omitempty"`

		Data map[string]string `json:"data,omitempty"`
	}

	type ConfigDataList struct {
		TypeMeta `json:",inline"`
		ListMeta `json:"metadata,omitempty"`

		Items []ConfigData `json:"items"`
	}
```

A new API resource selector (ResourceSelector) will be added to select objects and fields from those objects
for consuming API resources through the downward API.

``Registry`` interface will be added to accompany the ``ConfigData`` resource.  
``ConfigData`` information will be stored in ``etcd`` by default.

### Volume Source

A new ``ConfigDataVolumeSource`` type of volume source containing the ``ConfigData`` object will be added to the ``VolumeSource`` struct in the API:

```go
	type ConfigDataVolumeSource struct {
		ConfigDataName string `json:"configDataName"`
	}
```

**Note:** The update logic used in the downward API volume plug-in will be extracted and re-used in the volume plug-in for ``ConfigData``.

## Creating ConfigData
``ConfigData`` can be created in two ways explicitly and implicitly.

### Backwards-Compatibility
In order to smooth the transaction of component configuration from other means to ``ConfigData``
and preserve backwards-compatibility, user will be able to configure components the old way
and use an optional ``--write-config=<path>`` flag to create a ``ConfigData`` object containing
the given configuration options to ``<path>``, which can be created by invoking ``kubectl create -f <path>``.

For example, a ``kubelet`` will be configured as follows:
```bash
kubelet --address=0.0.0.0 --port=8888 --api-servers=10.240.13.14:8998,10.240.51.5:8998 --register-node=true --write-file=kubelet-config.yaml
```

The persisted object will have the following contents for the given configuration parameters:
```json
{
    "apiVersion": "v1",
    "kind": "ConfigData",
    "data": {
        "ADDRESS": "0.0.0.0",
        "PORT": "8888",
        "API_SERVERS": "10.240.13.14:8998, 10.240.51.5:8998",
        "REGISTER_NODE": "true",
    }
}
```

An identically configured kubelet can be created as such from now on:
```bash
kubelet --config-data=kubelet-config.yaml
```

The command-line flags given to the ``kubelet`` will be transformed to ``Data`` fields
inside a new ``ConfigData`` object for its configuration.

**Note:** Tooling will be implemented to facilitate this functionality for each
component that is going to migrate to ``ConfigData``.

### Explicit Creation
The following file ``pod-config-example.json`` contains data intended to be consumed as environment variables.
```json
{
    "apiVersion": "v1",
    "name": "pod-config-example",
    "kind": "ConfigData",
    "data": {
        "server_address": "10.240.13.14:8998",
        "db_address": "10.240.55.3",
        "db_port": "3306"
    }
}
```

To persist this ConfigData object to the API server, use the following command:
```bash
kubectl create -f pod-config.json
```

Now, the ``pod-config-example`` ConfigData object is stored in ``etcd``.
The new object can be accessed, referenced, used and updated through the API server.

## Configuration Workflow

The component configuration workflow is given below:

1. ``ConfigData`` object is created through the API server.
2. New components using the ``ConfigData`` object are created.
3. ``ConfigData`` object is retrieved using the downward API.
4. ``Data`` key/value pairs of ``ConfigData`` will be exposed as either environment variables (``EnvVarSource``) or as a new volume (``ConfigDataVolumeSource``).
5. API server is watched for modifications to the object.
6. After a modification, actions 3, 4, and 5 are repeated.

**Note**: Environment variables specified in ``ConfigData`` will override the environment variables
with the same name found in the system if consumed as environment variables.

## Examples

#### Consuming ConfigData as mounted volumes

``pod-config-volume`` is intended to be used as a mounted volume containing two files with their respective contents.
```json
{
    "apiVersion": "v1",
    "name": "pod-config-volume",
    "kind": "ConfigData",
    "data": {
        "component_config": "COMPONENT_CONFIG_STRING_CONTENTS",
        "component_metadata": "COMPONENT_METADATA_STRING_CONTENTS"
    }
}
```

Then, configuration objects are persisted to the API server as follows:
```bash
kubectl create -f pod-configuration-vol.json 
```

ExampleReplicationController starts 2 replicas that runs the `busybox` container.
In this use case, the ``ConfigData`` object will be mounted as a volume and contain
two files named "config.yaml" and "metadata.json".
```yaml
apiVersion: v1
kind: ReplicationController
metadata:
  name: ExampleReplicationController
spec:
  replicas: 2
  selector:
    app: busybox
  template:
    metadata:
      labels:
        app: busybox
    spec:
      containers:
        - name: busybox
          image: gcr.io/google_containers/busybox
          ports:
            - containerPort: 80
          volumeMounts:
            - name: config-data-volume
      volumes:
      - name: config-data-volume
        downwardAPI:
          items:
            - path: "config.yaml"
              resourceRef:
                from: pod-config-volume
                fieldPath: component_config
            - path: "metadata.json"
              resourceRef:
                from: pod-config-volume
                fieldPath: component_metadata
```

Now, we create the replication controller:
```bash
kubectl create -f example-replication-controller.yaml
```

New pods created by the ExampleReplicationController will use the ``pod-config-volume`` objects when starting/replicating pods.

``pod-config-volume`` object will be watched for any modifications and the mounted volumes will be updated if there are any modifications.

#### Consuming ConfigData as environment variables

```json
{
    "apiVersion": "v1",
    "name": "pod-config-env",
    "kind": "ConfigData",
    "data": {
        "api_server": "10.240.13.14:8998",
        "api_version": "v1",
        "db_host": "10.240.55.3",
        "db_port": "3306"
    }
}
```
``pod-config-env`` is intended to be consumed as environment variable key/value pairs.

We create our ``ConfigData`` object.
```bash
kubectl create -f pod-configuration-env.json
```

ExampleReplicationController starts 2 replicas that runs the `busybox` container.
In this use case, the ``ConfigData`` object will be consumed as 4 environment variables.
```yaml
apiVersion: v1
kind: ReplicationController
metadata:
  name: ExampleReplicationController
spec:
  replicas: 2
  selector:
    app: busybox
  template:
    metadata:
      labels:
        app: busybox
    spec:
      containers:
      - name: busybox
        image: gcr.io/google_containers/busybox
        ports:
        - containerPort: 80
        env:
        - name: API_SERVER
          valueFrom:
            resourceRef:
              from: pod-config-env
              key: api_servers
        - name: API_VERSION
          valueFrom:
            resourceRef:
              from: pod-config-env
              key: api_version
        - name: DB_HOST
          valueFrom:
            resourceRef:
              from: pod-config-env
              key: postgresql_host
        - name: DB_PORT
          valueFrom:
            resourceRef:
              from: pod-config-env
              key: postgresql_port
```

Now, we create the replication controller:
```bash
kubectl create -f example-replication-controller.yaml
```

New pods created by the ``ExampleReplicationController`` will use the ``pod-config-env`` objects when starting/replicating pods.

``pod-config-env`` object will be watched for any modifications and the currently consumed environment variables will be updated.

### Future Improvements

* Additional specification parameters can be added to the API for describing the expected behavior after
a modification to the consumed ``ConfigData`` object (e.g. 'on-restart: update').

### Considerations

* Multiple ``ConfigData`` objects can be supplied to the component for overriding or different consuming purposes (one for environment variables the other to be mounted as a volume)
