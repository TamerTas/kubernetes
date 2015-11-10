# Generic Configuration Object

## Abstract

This proposal proposes a generic configuration object named `ConfigData` stored inside `etcd` that
stores data used for the configuration of applications deployed on `Kubernetes`.

The main focus points of this proposal are:

* Dynamic distribution of configuration data to components and deployed applications.
* Encapsulate configuration information and simplify `Kubernetes` deployments.
* Create a flexible configuration model for `Kubernetes`.

## Motivation

Currently command-line flags and environment variables are used for configurations in `Kubernetes`
which is complicated and error-prone. This approach is not suitable for a distributed computing environment.
Specifically the problems with the current approach are:

* Complicated deployment orchestration
* Hard to synchronize across replicated components.
* Hard to change dynamically as needed by continuous services.

## Use Cases

1. Be able to load API resources as environment variables into a container.
2. Be able to load API resources as mounted volumes into a container.
3. Be able to update mounted volumes after a change to the API resource.

## Solution

A new `ConfigData` API resource is proposed to address these problems.

The `ConfigData` architecture will be very similar to that of `Secrets`.

`ConfigData` can either be mounted as a volume or can be injected as environment
variables to a container through the use of the downward API.

### Dynamic Configuration

Any long-running system has mutating specification over time. In order to facilitate this functionality,
`ConfigData` will be versioned and updates will automatically made available to the container.

`resourceVersion` (Found in `ObjectMetadata`) of the `ConfigData` object will be updated by the API server every time the object is modified.  After an update, modifications will be made visible to the consumer container. If the consumer uses the `Data` pairs only for initialization or during starting process, A rolling-update might be necessary to update the containers.

It is then the consumer container's responsibility to make use of the updated data
once it is made visible (or perform a rolling-update on consumers of that object. This is especially true if `ConfigData` is injected as environment variables. Dynamic configuration with environment variables is out of scope).

### Advantages

* Easy distribution through the API server.
* Provides configuration data in a consumer-agnostic manner.
* Persistent configuration information through API versioning.
* Ability to use `Kubernetes` authentication and security policies with configuration objects.
* By leveraging the power of `/watch` API or watching the volume, container processes can implement responsiveness to configuration changes.

## API Resource

A new resource for `ConfigData` will be added to the `Extentions` API Group:

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

type ConfigDataVolumeSource struct {
	ConfigDataName string `json:"configDataName"`
}

type ConfigDataSelector struct {
	APIVersion      string `json:"apiVersion, omitempty"`
	ConfigDataName  string `json:"from"`
	ConfigDataField string `json:"key"`
}
```

`Registry` interface will be added to "pkg/registry/configdata" along with the `ConfigData` resources.  

### Volume Source

A new `ConfigDataVolumeSource` type of volume source containing the `ConfigData` object will be added to the `VolumeSource` struct in the API:

```go
	type ConfigDataVolumeSource struct {
		ConfigDataName string `json:"configDataName"`
	}
```

**Note:** The update logic used in the downward API volume plug-in will be extracted and re-used in the volume plug-in for `ConfigData`.

## Creating ConfigData
The following file `pod-config-example.json` contains data intended to be consumed as environment variables.
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

Now, the `pod-config-example` ConfigData object is stored in `etcd`.
The new object can be accessed, referenced, used and updated through the API server.

## Configuration Workflow

### Environment Variables

The container configuration workflow is given below:

1. `ConfigData` object is created through the API server.
2. New containers using the `ConfigData` object are created.
3. `ConfigData` object is retrieved using the downward API.
4. `Data` key/value pairs of `ConfigData` will be exposed as environment variables (`EnvVarSource`).

### Mounted Volume

The container configuration workflow is given below:

1. `ConfigData` object is created through the API server.
2. New containers using the `ConfigData` object are created.
3. `ConfigData` object is retrieved using the downward API.
4. `ConfigData` will be mounted as a new volume (`ConfigDataVolumeSource`) where keys are file names and values are file contents.
5. API server is watched for modifications to the object.
6. After a modification, actions 3, 4, and 5 are repeated.

## Examples

#### Consuming ConfigData as volumes

`redis-volume-config` is intended to be used as a mounted volume containing two files with their respective contents.
```json
{
    "apiVersion": "extensions/v1beta1",
    "metadata": {
        "name": "redis-volume-config",
    },
    "kind": "ConfigData",
    "data": {
        "redis.conf": "pidfile /var/run/redis.pid\nport6379\ntcp-backlog 511\n databases 1\ntimeout 0\n",
        "slave_run.sh": "#!/bin/bash\nredis-server --slaveof ${REDISMASTER_SERVICE_HOST:-$SERVICE_HOST} $REDISMASTER_SERVICE_PORT\n",
        "redis_dockerbuild.sh": "docker build -t redis\ndocker build -t redis-master\ndocker build -t redis-slave\n"
    }
}
```

Configuration objects are persisted to the API server as follows:
```bash
kubectl create -f pod-configuration-vol.json 
```

ExampleReplicationController starts 2 replicas that runs the `busybox` container.
In this use case, the `ConfigData` object will be mounted as a volume and contain
four files named respectively "redis.conf", "slave_run.sh", and "redis_dockerbuild.sh".

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
              mountPath: /mnt/config-data
      volumes:
      - name: config-data-volume
        configDataVolumes:
          items:
            - path: "etc/redis.conf"
              configDataRef:
                from: redis-volume-config
                fieldPath: redis.conf
            - path: "scripts/slave_run.sh"
              configDataRef:
                from: redis-volume-config
                fieldPath: slave_run.sh
            - path: "scripts/redis_dockerbuild.sh"
              configDataRef:
                from: redis-volume-config
                fieldPath: redis_dockerbuild.sh
```

Now, we create the replication controller:
```bash
kubectl create -f example-replication-controller.yaml
```

New pods created by the ExampleReplicationController will use the `redis-volume-config` objects when starting/replicating pods.

The configuration volume presents an eventually-consistent view of the resource it consumes.
When a `ConfigData` consumed in a volume is updated, the contents of the volume will eventually change to hold the new values.
#### Consuming ConfigData as environment variables

```json
{
    "apiVersion": "extensions/v1beta1",
    "metadata": {
        "name": "etcd-env-config",
    },
    "kind": "ConfigData",
    "data": {
        "number_of_members": "3",
        "initial_cluster_state": "new",
        "initial_cluster_token": "DUMMY_ETCD_INITIAL_CLUSTER_TOKEN",
        "discovery_token": "DUMMY_ETCD_DISCOVERY_TOKEN",
        "discovery_url": "http://etcd-discovery:2379",
        "etcdctl_peers": "http://etcd:2379",
    }
}
```
`etcd-env-config` is intended to be consumed as environment variable key/value pairs.

We create our `ConfigData` object.
```bash
kubectl create -f etcd-env-config.json
```

The replication controller given below starts 2 replicas that runs the openshift-etcd container image.
In this use case, the `ConfigData` object will be consumed as 6 environment variables.

```yaml
apiVersion: extensions/v1beta1
kind: ReplicationController
metadata:
  name: ExampleReplicationController
spec:
  replicas: 3
  selector:
    name: etcd
  template:
    metadata:
      labels:
        name: etcd
    spec:
      containers:
      - name: etcd
        image: openshift/etcd-20-centos7
        ports:
        - containerPort: 2379
          protocol: TCP
        - containerPort: 2380
          protocol: TCP
        env:
        - name: ETCD_NUM_MEMBERS
          valueFrom:
            configDataRef:
              from: etcd-env-config
              key: num_members
        - name: ETCD_INITIAL_CLUSTER_STATE
          valueFrom:
            configDataRef:
              from: etcd-env-config
              key: initial_cluster_state
        - name: ETCD_DISCOVERY_TOKEN
          valueFrom:
            configDataRef:
              from: etcd-env-config
              key: dicovery_token
        - name: ETCD_DISCOVERY_URL
          valueFrom:
            configDataRef:
              from: etcd-env-config
              key: discovery_url
        - name: ETCDCTL_PEERS
          valueFrom:
            configDataRef:
              from: etcd-env-config
              key: etcdctl_peers
```

Now, we create the replication controller:
```bash
kubectl create -f example-replication-controller.yaml
```

Pods created by the `ExampleReplicationController` will use the `etcd-env-config` and environment variables
will be injected into the pods as specified.

### Future Improvements

* Additionally an init-container can be specified that watches `ConfigData` for modifications and restarts
the container processes consuming the particular `ConfigData` object.
