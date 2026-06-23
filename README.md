# File Downloader

[![Step changelog](https://shields.io/github/v/release/bitrise-steplib/steps-file-downloader?include_prereleases&label=changelog&color=blueviolet)](https://github.com/bitrise-steplib/steps-file-downloader/releases)

Download remote files to the build machine.

<details>
<summary>Description</summary>

The Step downloads selected files from a URL and places it where you want it to.

### Configuring the Step

1. Add a source to the **Download source URL** input. It can be either an Environment Variable or a URL.
1. Add a destination to the **Download destination path** input. The path should be relative to the root directory of the repository.

### Useful links

- [Using the File Downloader Step](https://devcenter.bitrise.io/tutorials/download-or-replace-project-resources/#using-the-file-downloader-step)
- [Managing files on Bitrise](https://devcenter.bitrise.io/getting-started/managing-files-on-bitrise/)

### Related Steps

- [Generate Text File](https://www.bitrise.io/integrations/steps/generate-text-file)
</details>

## 🧩 Get started

Add this step directly to your workflow in the [Bitrise Workflow Editor](https://docs.bitrise.io/en/bitrise-ci/workflows-and-pipelines/steps/adding-steps-to-a-workflow.html).

You can also run this step directly with [Bitrise CLI](https://github.com/bitrise-io/bitrise).

## ⚙️ Configuration

<details>
<summary>Inputs</summary>

| Key | Description | Flags | Default |
| --- | --- | --- | --- |
| `source` | The URL to download the file from. It can be either a direct URL or an Environment Variable that holds the URL. | required |  |
| `destination` | The path where the downloaded file will be saved. The path should be relative to the root directory of the repository. Any missing parent directories are created automatically, and a leading `~` is expanded to the home directory. | required |  |
| `file_permission` | Permission bits for the destination file in the format of `644`.  This option is useful for sensitive files that should have a permission like `600`.  When left empty, the file keeps the default permission set during download. |  |  |
</details>

<details>
<summary>Outputs</summary>
There are no outputs defined in this step
</details>

## 🙋 Contributing

We welcome [pull requests](https://github.com/bitrise-steplib/steps-file-downloader/pulls) and [issues](https://github.com/bitrise-steplib/steps-file-downloader/issues) against this repository.

For pull requests, work on your changes in a forked repository and use the Bitrise CLI to [run step tests locally](https://docs.bitrise.io/en/bitrise-ci/bitrise-cli/running-your-first-local-build-with-the-cli.html).

Learn more about developing steps:

- [Create your own step](https://docs.bitrise.io/en/bitrise-ci/workflows-and-pipelines/developing-your-own-bitrise-step/developing-a-new-step.html)
