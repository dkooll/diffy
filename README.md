# diffy

This test analyzes your terraform configurations and helps identify missing required properties and blocks, including nested dynamic ones.

## Notes

It retrieves dynamicly the schemas for the specified providers.

It filters out purely computed properties, which are typically populated by the provider.
