# Jellyfin on the K3s cluster

This application runs one Jellyfin server on a single-node K3s cluster. It stores server configuration under /k3s-volumes/jellyfin/config and reads TV series from a removable drive mounted at /media/bruno/KINGSTON.

## Prepare the media drive

The media hostPath uses type Directory, so the series directory must exist on the cluster host before the pod starts. The USB volume is mounted by the host's existing fstab configuration; this guide does not change it. Run these checks on the host and stop if the mount is absent or findmnt reports the wrong device:

    mountpoint -q /media/bruno/KINGSTON && findmnt --target /media/bruno/KINGSTON

Only after confirming the USB drive is mounted, create the destination and copy the series:

    sudo mkdir -p /media/bruno/KINGSTON/jellyfin/series
    sudo cp -r -- /path/to/your/series/. /media/bruno/KINGSTON/jellyfin/series/

The exFAT volume is root-owned on the host. Jellyfin runs as UID/GID 1000, so verify that this account can traverse the directories and read the media files using the drive's current mount permissions. Do not create or populate the series directory while the USB drive is unmounted; that would put files on the host filesystem under the mount point instead.

## What the manifests configure

The config PersistentVolume uses hostPath /k3s-volumes/jellyfin/config on the single-node cluster host and retains its data when the claim is released. The 1Gi PV/PVC capacity is Kubernetes storage accounting; hostPath does not reserve disk space or enforce a quota, so monitor free space on the host volume. The separate transcoding cache remains an emptyDir limited to 10Gi.

The container runs as UID/GID 1000 with fsGroup 1000. A root BusyBox init container changes ownership only on the mounted config directory and drops all capabilities except CHOWN. The media mount is read-only in the pod. Transcoding cache is an emptyDir mounted at /cache, limited to 10Gi and discarded whenever the pod is replaced.

The deployment uses an initial CPU request of 100m and memory request of 256Mi, with limits of 1 CPU and 2Gi memory. It also requests 1Gi and limits 12Gi of ephemeral storage. These are starting resource settings, not a promise of transcoding capacity. No GPU is configured, and 4K transcoding is not guaranteed.

The service is ClusterIP on port 80, and the Jellyfin ingress uses HTTP through Traefik's web entrypoint for `jellyfin.bsoapp.net`. The in-repository Caddy wildcard proxy covers only `*.internal.bsoapp.net`; it does not route this public hostname. The existing public Docker Cloudflare Tunnel sends traffic to Nginx, and Nginx has no Jellyfin route. The Kubernetes cloudflared deployment uses a tunnel token but contains no hostname or origin routing rules. Therefore this Ingress alone does not make Jellyfin publicly reachable: configure a Cloudflare Tunnel public-hostname route for `jellyfin.bsoapp.net` to an HTTP origin that reaches Traefik's web entrypoint (the existing Caddy configuration uses `192.168.0.44:32644` for that entrypoint). Ensure Cloudflare DNS for the hostname points through that tunnel. This route is configured outside these manifests and has not been verified. There is no NodePort, hostNetwork, or UDP discovery service, so clients should add the server URL manually.

## First setup and clients

After the app has been committed and pushed to the configured repository, the existing root Argo CD app watches the apps directory and should discover this Application. Its automated sync then reconciles the Jellyfin manifests. This is the expected GitOps flow; it does not mean the app has already synced.

After the public-hostname tunnel route is configured, open https://jellyfin.bsoapp.net from the Fire TV Stick or another client. Complete the initial setup and create the administrator account before inviting other users. Add a TV Shows library with the container path /media/series.

Use a clear series folder layout, for example:

    Example Show/
      Season 01/
        Example Show - S01E01.mp4

Include season and episode numbers such as S01E01 in each episode filename. In Jellyfin and on each client, set preferred audio and subtitle languages, then check the available tracks during playback.

On a Fire TV Stick, install the Jellyfin app, choose the option to add a server manually, and enter https://jellyfin.bsoapp.net. If the TV cannot connect, check that its DNS resolves this hostname and that the Cloudflare Tunnel public-hostname route reaches Traefik's HTTP entrypoint on port 32644. The TV's DNS and route have not been tested by these manifests.

Start with a small playback test. Check the playback details in Jellyfin to see whether the client is using direct play or transcoding. If transcoding reaches the CPU or memory limits, try a compatible media format or lower the playback quality; this deployment has no GPU acceleration and does not promise 4K transcoding.

## Backups and missing media

Verify that the existing host backup job includes the exact config directory /k3s-volumes/jellyfin/config and that a backup can be restored before relying on it. This change does not verify the job's configured data root or backup results. Jellyfin stores SQLite databases in its persistent config; use a backup method that captures a consistent copy. Before restoring the config, stop or quiesce Jellyfin through an Argo-aware maintenance procedure so it cannot write to the databases during the restore. Automated self-healing may revert a manual replica-count change. After restoring, start the app and check its logs. The /cache emptyDir is temporary and is not part of the config backup.

If the USB drive is removed or fails to mount, stop playback and pause library scans. Do not remove the library or treat its missing files as deletions. Reconnect the drive and confirm the expected device is mounted:

    mountpoint -q /media/bruno/KINGSTON && findmnt --target /media/bruno/KINGSTON

After the mount is restored, restart the deployment so its read-only bind mount attaches to the drive again, then confirm /media/series is populated before rescanning:

    kubectl -n production rollout restart deployment/jellyfin
    kubectl -n production exec deployment/jellyfin -- ls -la /media/series

For a separate service and pod diagnostic that bypasses Traefik, run this on a machine with kubectl access:

    kubectl -n production port-forward service/jellyfin 8096:80

While the command runs, open http://127.0.0.1:8096 on that same machine. This checks the Kubernetes service path from that machine; it does not test DNS or routing from the TV.
