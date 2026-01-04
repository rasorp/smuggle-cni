# Smuggle CNI
The Smuggle CNI Plugin is a meta plugin responsible for reading configuration
data written by the [Smuggle](https://github.com/rasorp/smuggle) agent and
delegating to the appropriate underlying CNI plugin to create the container's
network interface. The Smuggle CNI Plugin is expected to be installed on every
node in a cluster where Smuggle is used to provide networking.

### Docs
The documentation for Smuggle CNI is stored in the
[Smuggle Docs directory](https://github.com/rasorp/smuggle/tree/main/docs) along
with the rest of the Smuggle documentation.
