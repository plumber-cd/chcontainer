# chcontainer
Like chroot, but changes containers!

This is an experiment, not for production use.

See [`examples/basic`](examples/basic).

Basically, given the pod has required RBAC - it allows containers to `exec` into each other. Based on symlinks set in `CHC_SYMLINKS`, each container will create a symlink that in fact will run `chcontainer` so the target of a symlink will be actually executed in another container.

Theoretically this approach allows to make containers use each other, for example - `terraform` container might call `kubectl` or `helm` from within which are living in totally different containers.

Current prototype, however, needs to be added to every container. Perhaps via an `emptyDir` volume and an init-container? What if container is non-root? How to add `chcontainer` in `PATH`?
