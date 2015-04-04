# Generic Config Object

## Abstract

A generic ``Config`` object to store command-line flags and environment variables
that is used for configuration of ``Kubernetes`` components living on top of ``etcd``
is proposed in this proposal.

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
through the ``APIServer``. ``Config`` resource will be mounted as a new volume 
and will be managed by a volume plug-in that handles the ``APIServer`` communications.

### Advantages

* Reusable across different components.
* Easy distribution thanks to ``etcd``, a highly-available KVL store.
* By leveraging the power of ``/watch API`` along with the new resource, we can change component configurations dynamically.
* Layer of abstraction that gets rid of the global state currently used for command-line flags and environment variables.

### API Resource

A new resource for ``Config`` will be added to the ``API``:

```go
    type Config struct {
        TypeMeta   `json:",inline" yaml:",inline"`
        ObjectMeta `json:"metadata,omitempty" yaml:"metadata,omitempty"`

        // Flags contains command-line flags.
        Flags     map[string]string `json:"flags,omitempty" yaml:"flags"`

        // Variables contains environment variables.
        Variables map[string]string `json:"variables,omitempty" yaml:"flags"`
    }

    type ConfigList struct {
        TypeMeta `json:",inline" yaml:",inline"`
        ListMeta `json:"metadata,omitempty" yaml:"metadata,omitempty"`

        Items []Config `json:"items" yaml:"items"`
    }
```
``Registry`` interface will be added to accompany the ``Config`` resource.  
``Config`` information will be stored in ``etcd`` by default.

### Volume Plug-in

A new volume plug-in will be added to handle ``Config`` volumes.  
Plug-in will require access to the ``APIserver`` to retrieve the configuration
parameters. Responsibilities of the said plug-in will be:

* Retrieving ``Config`` from the ``APIServer``.
* Loading the configuration into the volume.

### Configuration

Configured component (with ``Kubectl`` or possibly other means) will send it's configuration to the ``APIServer``.

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

### Validation

``Flags`` and ``Variables`` are planned to be validated by type-checking at the moment for types provided
beforehand for planned configuration parameters (``maxPodNumber`` => ``int``).

### Limitations

1. Possibly requires a ``ConfigController`` for synchronization that watches the ``Config`` object
through the ``/watch API`` and restarts configured components when configuration resource is modified.

