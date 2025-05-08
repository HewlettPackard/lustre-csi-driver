# sbin/

## Description

Contains Cray-built `mount.lustre` binary, a user-space tool that gets invoked during the lustre mount operation.
It is assumed that the host kernel running the Lustre CSI driver container already has lustre client capabilities.
Since the host kernel is used by the CSI container, all that is needed in the container are any user-space tools that
get invoked during the `mount -t lustre <mgs_nid_list> <dest>` operation. The only user-space tool used by the `mount
-t lustre ...` operation is `mount.lustre`; all other lustre client functionality resides in kernel-space (again, shared
by the host OS and the container).

## Where did this binary come from?

Cray internally builds and packages lustre-client RPMs with supporting functionality for a diverse range of networking
interfaces, including Ethernet (`@tcp` NID), Infiniband (`@o2ib` NID), and Slingshot (`@kfi` NID). This `mount.lustre`
is built to support parsing of a wider range of NID formats than what the publicly available lustre-client RPMs provide.

The source tree for lustre-client is found at https://github.com/cray/lustre/tree/cray-2.15.B21.

