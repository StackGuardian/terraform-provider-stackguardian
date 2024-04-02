
# Examples of provider usage

## Instructions for quickly testing out the provider

```shell
# First, from the project root, enter one of the example directories
cd examples/role_example

# Then clean, initialize and run Terraform to create and destroy the defined resource in the example
rm -rf .terraform .terraform.lock.hcl; pushd ../..; make install; popd; terraform init && terraform plan && terraform apply -auto-approve && terraform destroy --auto-approve
```
