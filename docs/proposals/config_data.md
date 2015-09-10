# Generic Configuration Object

## Abstract

This proposal proposes a generic configuration object named ``ConfigData`` stored inside ``etcd`` that
stores data used for the configuration of ``Kubernetes`` components.

The main focus of this proposal to solve dynamic configuration problem of components
and encapsulate configuration information to decrease component configuration complexity, 
to increase overall system quality, and to create a flexible configuration model for ``Kubernetes`` components.

## Motivation

Currently command-line flags and environment variables are used for component configurations in ``Kubernetes``
which is complicated and error-prone. This approach is not suitable for a distributed computing environment.
Specifically the problems with the current approach are:

* Complicated deployment orchestration
* Hard to synchronize across replicated components.
* Hard to change dynamically as needed by continuous services.
* Hard to change in different versions that may support different configuration options.

## Solution

A new ``ConfigData`` resource is proposed to solve the aforementioned problems
that is stored in ``etcd`` and distributed to interested/necessary components
through the API server. 

The ``ConfigData`` architecture will be very similar to that of ``Secrets``.

``ConfigData`` can either be mounted as a volume or can be injected as environment
variables to a container through the use of the downward API.

### Dynamic Configuration

Any long-running system has mutating specification over time. In order to facilitate this functionality,
``ConfigData`` will be versioned and updates will automatically made available to the container (downward API).

``resourceVersion`` (Found in ``ObjectMetadata``) of the ``ConfigData`` object will be updated by the server every time the object is modified.
After an update, modifications will be made visible to the consumer container. If the consumer uses the ``Data`` pairs
only for initialization or during starting process, A rolling-update might be necessary to update the components.

It is then the consumer component's responsilibity to make use of the updated data
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
``Registry`` interface will be added to accompany the ``ConfigData`` resource.  
``ConfigData`` information will be stored in ``etcd`` by default.

### Volume Source

A new ``ConfigDataSource`` type of volume source containing the ``ConfigData`` object will be added to the ``VolumeSource`` struct in the API:

```go
	type ConfigDataVolumeSource struct {
		ConfigDataName string `json:"configDataName"`
	}
```

This volume source will be made available and updated by using the downward API volume plug-in (might require some generalizations).


## Creating ConfigData
``ConfigData`` can be created in two ways explicitly and implicitly.

### Backwards-Compatibility
In order to smooth the transaction of component configuration from other means to ``ConfigData``
and preserve backwards-compatibility, user will be able to configure components the old way
which is going to be transformed into a ``ConfigData`` object implicitly which can be updated and reused.

For example, a ``kubelet`` is configured currently as follows:
```bash
kubelet --address=0.0.0.0 --port=8888 --api-servers=10.240.13.14:8998,10.240.51.5:8998 --register-node=true
```

After this proposal is integrated, when you invoke the same commands,
a ``ConfigData`` object will be created in the API server.
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
# KubeletConfigurationRef is not clear whether it will be automatically
# generated or it will be a label selector (e.g. use the config-data of component X).
kubelet --config-data=KubeletConfigurationRef
```

The command-line flags given to the ``kubelet`` will be transformed to ``Data`` fields
inside a new ``ConfigData`` object for its configuration.

**Note:** Tooling will be implemented to facilitate this functionality for each
component that is going to migrate to ``ConfigData``.

### Explicit Creation
The following file ``pod-config.json`` contains data intended to be consumed as environment variables.
```json
{
    "apiVersion": "v1",
    "name": "PodConfiguration",
    "kind": "ConfigData",
    "data": {
        "API_SERVER": "10.240.13.14:8998",
        "API_VERSION": "v1",
        "DB_HOST": "10.240.55.3",
        "DB_PORT": "3306"
    }
}
```

To persist this ConfigData object to the API server, use the following command:
```bash
kubectl create -f pod-config.json
```

Now, the ``PodConfiguration`` ConfigData object is stored in ``etcd``.
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

## Use Cases

#### Pod Configuration

```json
{
    "apiVersion": "v1",
    "name": "PodConfiguration1",
    "kind": "ConfigData",
    "data": {
        "API_SERVER": "10.240.13.14:8998",
        "API_VERSION": "v1",
        "DB_HOST": "10.240.55.3",
        "DB_PORT": "3306"
    }
}
```
``PodConfiguration1`` is intended to be consumed as environment variable key/value pairs.

```json
{
    "apiVersion": "v1",
    "name": "PodConfiguration2",
    "kind": "ConfigData",
    "data": {
        "config.yaml": "CONFIG_YAML_STRING_CONTENTS",
        "spec.yaml": "SPEC_YAML_STRING_CONTENTS"
    }
}
```
``PodConfiguration2`` is intended to be used as a mounted volume containing two files with their respective contents.

Then, configuration objects are persisted to the API server as follows:
```bash
kubectl create -f pod-configuration-1.json 
kubectl create -f pod-configuration-2.json
```

We create a replication controller specification that creates pods using these two configuration objects:
```yaml
apiVersion: v1
kind: ReplicationController
metadata:
  name: ExampleReplicationController
spec:
  replicas: 2
  selector:
    app: dummy-app
  template:
    metadata:
      labels:
        app: dummy-app
    spec:
      # New pods will be created using the ConfigData objects.
      # Each config object will be updated be it a volume or environment variables.
      # But, some of them might require a restart on update.
      # Note: If multiple env ConfigData objects are specified, let's call them Env1 and Env2.
      #       If both of them contains a same named environment variable, Env2 will be the final one.
      #       If Env1 is updated, the reload process will be Env1 first, then Env2.
      #       Even though Env2 is not updated, the 'env' ConfigData objects are reloaded in the
      #       order they are specified to prevent Env1 from overriding any of the variables in Env2.
      # TODO: 'config-data' is a separate field because there are different ways to consume the object
      #       and 'volume' might be a bit too specific. More discussions on this will be needed.
      config-data:
          env:
          # PodConfiguration1 is loaded as environment variables into the
          # main container of the pod. It requires a restart when the object is updated
          - name: PodConfiguration1
            on_update: restart
          volume:
          # PodConfiguration2 is mounted as a new volume to the main container of the pod.
          - name: PodConfiguration2
      containers:
      - name: dummy-app
        image: dummy-app
        ports:
        - containerPort: 80
```

example-replication-controller starts 2 replicas that runs the `dummy-app` container.
Pod template states how these pods will make use of the ``ConfigData`` objects.

Now, we create the replication controller:
```bash
kubectl create -f example-replication-controller.yaml
```

New pods created by the ExampleReplicationController will use the ``PodConfiguration`` objects when starting/replicating pods.

When the ``PodConfiguration`` objects are modified the new object will be retrieved and loaded into the component.

### Considerations

* ``ConfigData`` object might easily be extended to include other configuration components used in ``Kubernetes`` to configure it's behaviour one way or another (e.g. config files).
* Multiple ``ConfigData`` objects can be supplied to the component for overriding or different consuming purposes (one for environment variables the other to be mounted as a volume)
