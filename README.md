<a href="https://www.stackguardian.io/">
    <img src=".github/stackguardian_logo.svg" alt="StackGuardian logo" title="StackGuardian" align="right" height="40" />
</a>

# StackGuardian Terraform Provider

_**DISCLAIMER:** This Terraform Provider project is currently in pre-release mode and is intended to be used with StackGuardian test servers and not in production environments_.

The [StackGuardian Terraform Provider](https://github.com/StackGuardian/terraform-provider-stackguardian) allows [Terraform](https://www.terraform.io/) to programmatically interact with the [StackGuardian Orchestrator Test API](https://docs.qa.stackguardian.io/docs/api/overview) to help you manage resources on StackGaudian platform and ultimatelty enabling organizations to manage cloud infrastructure in a cost-efficient, secure, and compliant way.

The [StackGuardian Terraform Provider](https://github.com/StackGuardian/terraform-provider-stackguardian) allows [Terraform](https://www.terraform.io/) to programmatically interact with the [StackGuardian Orchestrator Test API](https://docs.qa.stackguardian.io/docs/api/overview) to help you manage resources on StackGaudian platform and ultimatelty enabling organizations to manage cloud infrastructure in a cost-efficient, secure, and compliant way.

## Documentation

This Terraform provider currently supports the following StackGuardian resources:

- Connector (Cloud and Version Control)
- Workflow Group
- Role
- Role Assignment (Users and SSO Groups)

Please refer to the [onboarding examples files](/docs-guides-assets/onboarding) for details on how to work with these resources. Other resources like Policies, Runner Groups etc. are work under progress and will be released in the future releases. You can show your interested in new features by creating issues in our GitHub repo.

## Installation steps

**This version of the StackGuardian Terraform provider is currently in pre release and is not yet available on the Terraform Registry.**\
To install it, you will need to download the zip file for your platform and architecture from the [latest pre release](https://github.com/StackGuardian/terraform-provider-stackguardian/releases/tag/v1.0.0-rc) and extract it to your local machine.

1. Create a directory for the StackGuardian Terraform provider.

   **Linux/MacOS**\
   Replace `<OS_ARCH>` in the command below with the operating system and architecture of your machine, for example `linux_amd64`, `darwin_arm64`.

   ```
   export OS_ARCH=<OS_ARCH>
   mkdir -p ~/.terraform.d/plugins/terraform.local/local/StackGuardian/1.0.0-rc/${OS_ARCH}
   ```

   **Windows**

   ```
   mkdir %USERPROFILE%\.terraform.d\plugins\terraform.local\local\StackGuardian\1.0.0-rc\windows_amd64
   ```

2. We need to make sure that the following configuration is set in the `.terraformrc` file so that Terraform can find the local StackGuardian Terraform provider.\
   _Please replace `<Fully qualified path to .terraform.d/plugins>` with the fully qualified path to your `.terraform.d/plugins` directory._
   `    provider_installation 
    {
        filesystem_mirror 
        {
            path    = "<Fully qualified path to .terraform.d/plugins>"
        }
        direct 
        {
            exclude = ["terraform.local/*/*"]
        }
    }
   `

3. Download the zip file for your platform and architecture from the [latest pre release](https://github.com/StackGuardian/terraform-provider-stackguardian/releases/tag/v1.0.0-rc) and extract it to the directory you created in step 1.\
   **Linux/MacOS**

   ```
   curl -L -o ~/.terraform.d/plugins/terraform.local/local/StackGuardian/1.0.0-rc/${OS_ARCH}/terraform-provider-stackguardian_v1.0.0-rc.tar.gz \
   "https://github.com/StackGuardian/terraform-provider-stackguardian/releases/download/v1.0.0-rc/terraform-provider-stackguardian_${OS_ARCH}.tar.gz"
   ```

   ```
   cd ~/.terraform.d/plugins/terraform.local/local/StackGuardian/1.0.0-rc/${OS_ARCH}
   tar -xvf terraform-provider-stackguardian_v1.0.0-rc.tar.gz
   ```

   **Windows**
   _On Windows using curl and tar requires Windows 10 version 1803 or later._

   ```
   curl -L -o terraform-provider-stackguardian_v1.0.0-rc.zip \
   "https://github.com/StackGuardian/terraform-provider-stackguardian/releases/download/v1.0.0-rc/terraform-provider-stackguardian_Windows_x86_64.zip"
   ```

   ```
   tar -xfz terraform-provider-stackguardian_v1.0.0-rc.zip -C %USERPROFILE%\.terraform.d\plugins\terraform.local\local\StackGuardian\1.0.0-rc\windows_amd64\
   ```

4. [Optional] Delete the downloaded zip file.\
   **Linux/MacOS**

   ```
   rm terraform-provider-stackguardian_v1.0.0-rc.tar.gz
   ```

   **Windows**

   ```
   del terraform-provider-stackguardian_v1.0.0-rc.zip
   ```

5. For Linux and MacOS systems you might need to set execute permissions on the terraform-provider-stackguardian binary.

   ```
   chmod +x terraform-provider-stackguardian
   ```

6. All done! You can now use the StackGuardian Terraform provider to create and manage resources in your StackGuardian test environment organization.

To get started you can try `project-01` and `project-02` from the [onboarding examples files](/docs-guides-assets/onboarding).
_Do remember to replace the `<ORG_NAME>` and `<API_KEY>` placeholders in the provider definition with your actual values from your organization on the StackGuardian test environment._

## Contributing

This project is currently only open to limited external contributions, please reachout to [@akshat0694](https://github.com/akshat0694).
It will become generalliy available for contrinutions after the release of v1.0.0.
