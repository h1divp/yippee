<script lang="ts">
    import { goto } from '$app/navigation';
    import { user } from '$lib/stores/auth';

    $: if (!$user) {
        goto('/login');
    }
</script>

<slot/>