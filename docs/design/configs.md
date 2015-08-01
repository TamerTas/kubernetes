# Generic Config Object

## Abstract

A generic ``Config`` object to store command-line flags and environment variables
that is used for configuration of ``Kubernetes`` components living on top of ``etcd``
is proposed in this proposal.

The main focus of this proposal to solve dynamic configuration problem of components
and encapsulate configuration information to increase overall system quality and to
create a flexible configuration model for ``Kubernetes`` components.

## Problem

Currently command-line flags and environment variables are used for component configurations in ``Kubernetes``.
But, this approach is not suitable for a distributed computing environment. Specifically the
problems with the current approach are:

* Hard to synchronize across replicated components.
* Hard to change dynamically as needed by continuous services.
* Hard to change in different versions that may support different configuration options.

## Solution

A new ``Config`` resource is proposed to solve the aforementioned problems
that is stored in ``etcd`` and distributed to interested/necessary components
through the API server. ``Config`` resource will be mounted as a new volume
and will be managed by a side-car container that communicates with the API server.

The ``Config`` resource architecture will be very similar to that of ``Secrets``.
However, the ``Config`` resource will be dynamic, meaning that different components consuming
the same ``Config`` will be notified about the changes to that particular resource.

### Advantages

* Reusable across different components.
* Easy distribution through the API server.
* By leveraging the power of ``/watch API`` along with the new resource, we can change component configurations dynamically.
* Layer of abstraction that gets rid of the global state currently used for command-line flags and environment variables.

### API Resource

A new resource for ``Config`` will be added to the ``API``:

```go
    type ConfigType int
    const (
        ConfigTypeOpaque ConfigType = 1 << iota
        ConfigTypeEnvVar
        ConfigTypeFlag
        ConfigTypeBLOB
    )

	type Config struct {
		TypeMeta   `json:",inline"`
		ObjectMeta `json:"metadata,omitempty"`

        Data map[string]string `json:"data,omitempty"`

        Type ConfigType `json:"type,omitempty"`
	}

	type ConfigList struct {
		TypeMeta `json:",inline"`
		ListMeta `json:"metadata,omitempty"`

		Items []Config `json:"items"`
	}
```
``Registry`` interface will be added to accompany the ``Config`` resource.  
``Config`` information will be stored in ``etcd`` by default.

### Volume Source

A new ``ConfigSource`` type of volume source containing the ``Config`` object will be added to the ``VolumeSource`` struct in the API:

```go
	type ConfigVolumeSource struct {
        ConfigName string `json:"configName"`
	}
```

### Configuration

Configured component (with ``Kubectl`` or possibly other means) will send it's configuration to the API server.

```json
{
    "apiVersion": "v1beta3",
    "kind": "Config",
    "id": "",
    "flags": {
        "api_server": "132.32.13.14:8998",
        "api_version": "v1beta3",
    },
    "variables": {
        "DB_HOST": "1.1.1.1",
        "DB_PORT": "8778",
    }
}
```

The owner of the configuration will be able mutate the configuration and it will be immutable from the
perspective of other components (**Note:** A ``ConfigController`` might be added to undertake this responsibility).

Consumers will have a side-car container watching the API for changes to `Config`
and reload the `Config` after a change.

After the `Config` is loaded the data will be visible inside the main container.
The application will then be responsible from using the data.

### Dynamic Configuration

A side-car container will watch the API server and update the ``ConfigVolumeSource``.
It will make the configuration visible inside the container and propagate updates to
the API server if needs be.

The said container's responsibilities will be as follows:

* Retrieving ``Config`` from the API server.
* Loading the configuration into the volume.
* Sending the resource to the API server.

### Limitations

1. Might require a ``ConfigController`` for synchronization that restarts configured components when configuration resource is modified.

### Considerations

* ``Config`` object might easily be extended to include other configuration components used in ``Kubernetes``
 to configure it's behaviour one way or another (e.g. config files).
* Plug-in architecture might be used instead of side-car containers if long-running plug-in support is added.
