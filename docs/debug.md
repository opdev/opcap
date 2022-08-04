# Debugging Opcap in VS Code
### Prerequisite
  Access to Single Node Openshift (SNO) Cluster

## The following steps will help you setup your debugging environment

1.  Append the following lines to launch.json file. (If no launch.json has been created, VS Code shows the Run start view when you Select Run and Debug or press F5)

    ```text
    // Using the --filter-packages flag to audit a single operator,
    "args": ["check", "--filter-packages=aikit-operator"],
    // or use no flags to audit the entire packagemanifests,
    "args": ["check"],
    // Specify the path where kubeconfig file is located
    "env": {"KUBECONFIG":"/Full-path-to-kubeconfig-file"}
    ```

3. Add breakpoints to the code
4. Select main.go
5. Select Run and Debug or press F5

## Tips

* A good starting point is to add a breakpoint to the OperatorInstall() function in the operator_install.go file.

* If you want to **test Operand install**, you need to change the defaultAuditPlan variable in the check.go file, and add a breakpoint to the OperandInstall() function in the operand_install. We support --auditplan flag, but it hasn't been turned on to accept values.
```text
var defaultAuditPlan = []string{"OperatorInstall", "OperandInstall", "OperandCleanUp", "OperatorCleanUp"}
```
