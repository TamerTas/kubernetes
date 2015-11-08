# Generic Configuration Object

## Abstract

This proposal proposes a generic configuration object named ``ConfigData`` stored inside ``etcd`` that
stores data used for the configuration of applications deployed on ``Kubernetes``.

The main focus points of this proposal are:

* Dynamic distribution of configuration data to components and deployed applications.
* Encapsulate configuration information and simplify ``Kubernetes`` deployments.
* Create a flexible configuration model for ``Kubernetes``.

## Motivation

Currently command-line flags and environment variables are used for configurations in ``Kubernetes``
which is complicated and error-prone. This approach is not suitable for a distributed computing environment.
Specifically the problems with the current approach are:

* Complicated deployment orchestration
* Hard to synchronize across replicated components.
* Hard to change dynamically as needed by continuous services.

## Use Cases

1. Be able to load API resources as environment variables into a component.
2. Be able to load API resources as mounted volumes into a component.
3. Be able to update mounted volumes after a change to the API resource.

## Solution

A new ``ConfigData`` API resource is proposed to address these problems.

The ``ConfigData`` architecture will be very similar to that of ``Secrets``.

``ConfigData`` can either be mounted as a volume or can be injected as environment
variables to a container through the use of the downward API.

### Dynamic Configuration

Any long-running system has mutating specification over time. In order to facilitate this functionality,
``ConfigData`` will be versioned and updates will automatically made available to the container.

``resourceVersion`` (Found in ``ObjectMetadata``) of the ``ConfigData`` object will be updated by the API server every time the object is modified.  After an update, modifications will be made visible to the consumer container. If the consumer uses the ``Data`` pairs only for initialization or during starting process, A rolling-update might be necessary to update the components.

It is then the consumer component's responsibility to make use of the updated data
once it is made visible (or perform a rolling-update on consumers of that object. This is especially true if ``ConfigData`` is injected as environment variables. Dynamic configuration with environment variables is out of scope).

### Advantages

* Easy distribution through the API server.
* Provides configuration data in a consumer-agnostic manner.
* Persistent configuration information through API versioning.
* Ability to use ``Kubernetes`` authentication and security policies with configuration objects.
* By leveraging the power of ``/watch`` API or watching the volume, container processes can implement responsiveness to configuration changes.

## API Resource

A new resource for ``ConfigData`` will be added to the ``Extentions`` API Group:

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

A new API resource selector (``ConfigDataSelector``) will be added to select objects and fields from those objects
for consuming API resources through the downward API.

``Registry`` interface will be added to accompany the ``ConfigData`` resource.  

### Volume Source

A new ``ConfigDataVolumeSource`` type of volume source containing the ``ConfigData`` object will be added to the ``VolumeSource`` struct in the API:

```go
	type ConfigDataVolumeSource struct {
		ConfigDataName string `json:"configDataName"`
	}
```

**Note:** The update logic used in the downward API volume plug-in will be extracted and re-used in the volume plug-in for ``ConfigData``.

## Creating ConfigData
The following file ``pod-config-example.json`` contains data intended to be consumed as environment variables.
```json
{
    "apiVersion": "extensions/v1beta1",
    "metadata": {
        "name": "pod-config-example",
    },
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

### Environment Variables

The component configuration workflow is given below:

1. ``ConfigData`` object is created through the API server.
2. New components using the ``ConfigData`` object are created.
3. ``ConfigData`` object is retrieved using the downward API.
4. ``Data`` key/value pairs of ``ConfigData`` will be exposed as environment variables (``EnvVarSource``).
5. API server is watched for modifications to the object.
6. After a modification, actions 3, 4, and 5 are repeated.

### Mounted Volume

The component configuration workflow is given below:

1. ``ConfigData`` object is created through the API server.
2. New components using the ``ConfigData`` object are created.
3. ``ConfigData`` object is retrieved using the downward API.
4. ``ConfigData`` will be mounted as a new volume (``ConfigDataVolumeSource``) where keys are file names and values are file contents.
5. API server is watched for modifications to the object.
6. After a modification, actions 3, 4, and 5 are repeated.

## Examples

#### Consuming ConfigData as mounted volumes

``pod-config-volume`` is intended to be used as a mounted volume containing two files with their respective contents.
```json
{
    "apiVersion": "extensions/v1beta1",
    "metadata": {
        "name": "pod-config-volume",
    },
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
apiVersion: extensions/v1beta1
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
        configDataVolumes:
          items:
            - path: "config.yaml"
              configDataRef:
                from: pod-config-volume
                fieldPath: component_config
            - path: "metadata.json"
              configDataRef:
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
    "apiVersion": "extensions/v1beta1",
    "metadata": {
        "name": "pod-config-env",
    },
    "kind": "ConfigData",
    "data": {
        "api_server": "10.240.13.14:8998",
        "api_version": "extensions/v1beta1",
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
apiVersion: extensions/v1beta1
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
            configDataRef:
              from: pod-config-env
              key: api_servers
        - name: API_VERSION
          valueFrom:
            configDataRef:
              from: pod-config-env
              key: api_version
        - name: DB_HOST
          valueFrom:
            configDataRef:
              from: pod-config-env
              key: postgresql_host
        - name: DB_PORT
          valueFrom:
            configDataRef:
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
