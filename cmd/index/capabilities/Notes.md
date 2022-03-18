# Notes

1. create secret with kubeconfig with name `kube-config-secret` in `default` namespace
2. create secret with dockerconfig to access private registries(registrey.connect.redhat.com, registry.redhat.io) with name `registry-redhat-dockerconfig` in `default` namespace
3. Make sure that the `default` serviceAcctount has the secrets created above
4. Run the k8s job with the updated audit-tool image (quay.io/opdev/audit-tool:v0.0.7)
    ```
    oc create -f job-audit-tool.yaml
    ```


(only for non-default namespace) Add scc policy to default serviceaccount in `openshift-operator-lifecycle-manager` namespace by running:`oc adm policy add-scc-to-user privileged -z default -n openshift-operator-lifecycle-manager`