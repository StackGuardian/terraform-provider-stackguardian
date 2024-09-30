<a href="https://www.stackguardian.io/">
    <img src=".github/stackguardian_logo.svg" alt="StackGuardian logo" title="StackGuardian" align="right" height="40" />
</a>

# StackGuardian Terraform Provider


_**DISCLAIMER:** This Terraform Provider project is currently in pre-release mode and is intended to be used with StackGuardian QA servers and not in production environments_.

The [StackGuardian Terraform Provider](https://github.com/StackGuardian/terraform-provider-stackguardian) allows [Terraform](https://www.terraform.io/) to programmatically interact with the [StackGuardian Orchestrator](https://docs.stackguardian.io/) [API](https://docs.stackguardian.io/docs/api/overview) to help you manage your cloud infrastructure in a cost-efficient, secure, and compliant way.



## Documentation

This Terraform provider currently supports the following StackGuardian resources: 
- Connector
- Workflow Group
- Role
- Role Assignment (Users)

Please refer to the [onboarding examples files](/docs-guides-assets/onboarding) for details on how to work with these resources.

## Installation steps

**This version of the StackGuardian Terraform provider is currently in pre release and is not yet available on the Terraform Registry.**\
To install it, you will need to download the zip file for your platform and architecture from the [latest release](https://github.com/StackGuardian/terraform-provider-stackguardian/releases/latest) and extract it to your local machine.

_**Note:** Replace `<OS_ARCH>` in the commands below with the operating system and architecture of your machine, for example `linux_amd64`, `darwin_arm64` or `windows_amd64`._

1. Create a directory for the StackGuardian Terraform provider.\
    **Linux/MacOS**
    ```
    mkdir -p ~/.terraform.d/plugins/local/StackGuardian/v1.0.0-rc/<OS_ARCH>

    ```
    **Windows**
    ```
    mkdir %USERPROFILE%\.terraform.d\plugins\local\StackGuardian\v1.0.0-rc\<OS_ARCH>
    ```
2. Download the zip file for your platform and architecture from the [latest release](https://github.com/StackGuardian/terraform-provider-stackguardian/releases/latest) and extract it to the directory you created in step 1.\
    _On Windows using curl and tar requires Windows 10 version 1803 or later._
    ```
    curl -L -o terraform-provider-stackguardian_v1.0.0-rc.zip \
    "https://github.com/StackGuardian/terraform-provider-stackguardian/releases/download/v1.0.0-rc/terraform-provider-stackguardian_1.0.0-rc_<OS_ARCH>.zip"
    ```
    **Linux/MacOS**
    ```
    unzip terraform-provider-stackguardian_v1.0.0-rc.zip -d ~/.terraform.d/plugins/local/StackGuardian/v1.0.0-rc/<OS_ARCH>
    ``` 
    **Windows**
    ```
    tar -xf terraform-provider-stackguardian_v1.0.0-rc.zip -C %USERPROFILE%\.terraform.d\plugins\local\StackGuardian\v1.0.0-rc\<OS_ARCH>
    ```

3. [Optional] Delete the downloaded zip file.\
    **Linux/MacOS**
    ```
    rm terraform-provider-stackguardian_v1.0.0-rc.zip
    ```
    **Windows**
    ```
    del terraform-provider-stackguardian_v1.0.0-rc.zip
    ```

4. For Linux and MacOS systems you might need to set execute permissions on the terraform-provider-stackguardian binary.
    ```
    chmod +x ~/.terraform.d/plugins/local/StackGuardian/v1.0.0-rc/<OS_ARCH>/terraform-provider-stackguardian
    ```

5. All done! You can now use the StackGuardian Terraform provider to create and manage resources in your StackGuardian QA account.

To get started you can try `project-01` and `project-02` from the [onboarding examples files](/docs-guides-assets/onboarding).
_Do remember to replace the `<ORG_NAME>` and `<API_KEY>` placeholders in the provider definition with your actual values from your account on the StackGuardian QA environment._

## Contributing
This project is currently not able to accept external contributions.
It will become possible after releasing a stable version later on.
