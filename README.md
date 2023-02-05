# KubeFixtures
KubeFixtures is a CLI tool/library to help you load fixtures(status included) into your Kubernetes cluster.


## Why?
When you're working in a team and CRD has been defined but haven't implemented the operator yet, you can write up couple sample YAML files with spec and status and load them into your cluster. This will help others work on components(especially APIs that rely on kubernetes control plane) that depend on the CRD while the operator is under development.


## Example workflow
- Write up a YAML file with spec and status
- Run `kubefixtures load -f <yaml file>`
- Run `kubectl get ..` to view the applied resource. 
- Create a new file with the updated status to mimick transition to the new state
- Run `kubefixtures transitition -f <yaml file>` to update the status of the resource
- Develop your component for particular states


## FAQ
- How is this different from `kubectl apply -f`?

Kubectl will only apply the spec part of the YAML file. KubeFixtures will apply the entire YAML file. This is useful when you want to load a YAML file with spec and status.

- Is there any other way to patch status than use this tool?

You could run `curl` commands to patch status but that can get tiresome. This tool is meant to make it easier to declartiively update current state into your cluster.

- What's the minimalistic way to get started?

You can get started by using any Kubernetes cluster. If you're looking to be even more minimalistic, you can use [kcp](https://github.com/kcp-dev/kcp) as no workloads need to be running on the cluster. Once you have a kubeconfig file, you can provide `--kubconfig` flag or set an env variable `KUBECONFIG` to the path of the kubeconfig file into your cluster.

## Contributing
Please feel free to open issues and PRs. 