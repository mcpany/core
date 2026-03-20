# Tag-based Access Control

The **Tag-based Access Control** feature allows you to restrict service access for user profiles based on tags. This provides a scalable way to manage permissions, ensuring that users only have access to the services they need.

## Overview

In the Profile Editor, you can now define **Access Tags** in addition to the standard Environment Type (Dev, Prod, Debug). Services that have matching tags are automatically included in the profile's allowed list.

![Profile Editor with Tags](../screenshots/profile_editor.png)

## How it Works

1.  **Tag Services**: Assign tags to your services (e.g., `finance`, `hr`, `eu-region`) when configuring them.
2.  **Configure Profiles**: In the Profile Editor, add the corresponding Access Tags.
3.  **Automatic Selection**: Any service that matches one of the profile's tags will be automatically enabled for that profile.
4.  **Visual Feedback**: Auto-selected services are marked with an `Auto` badge and the checkbox is disabled to indicate it's enforced by the tag rule.
5.  **Explicit Overrides**: You can still manually select additional services that don't match the tags.

## Benefits

-   **Scalability**: Manage permissions for hundreds of services by tagging them, rather than updating every profile individually.
-   **Security**: Ensure new sensitive services (e.g., `finance`) are not accidentally exposed to general profiles.
-   **Flexibility**: Combine role-based (via tags) and instance-based (explicit selection) access control.
