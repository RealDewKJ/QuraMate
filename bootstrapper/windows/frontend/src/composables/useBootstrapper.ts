import { computed, onMounted, ref } from 'vue';

type BootstrapperPhase = 'loading' | 'ready' | 'installing' | 'handoff' | 'error';

interface StatusPayload {
  stage?: string;
  message?: string;
}

const fallbackContext: BootstrapperLaunchContext = {
  mode: 'install',
  version: 'dev',
  installerPath: '',
  executablePath: '',
  installDir: 'C:\\Program Files\\QuraMate',
};

export function useBootstrapper() {
  const context = ref<BootstrapperLaunchContext>(fallbackContext);
  const phase = ref<BootstrapperPhase>('loading');
  const message = ref('Getting ready');
  const error = ref('');

  const primaryActionLabel = computed(() => {
    if (phase.value === 'installing' || phase.value === 'handoff') {
      return 'Installing';
    }
    return context.value.mode === 'update' ? 'Update QuraMate' : 'Install QuraMate';
  });

  const canStart = computed(() => phase.value === 'ready');

  const loadContext = async () => {
    try {
      const app = window.go?.main?.App;
      if (!app) {
        phase.value = 'ready';
        message.value = 'Preview mode';
        return;
      }

      context.value = await app.GetLaunchContext();
      phase.value = 'ready';
      message.value = 'Ready to install';
    } catch (err) {
      phase.value = 'error';
      error.value = err instanceof Error ? err.message : 'Unable to load bootstrapper context';
    }
  };

  const startInstall = async () => {
    const app = window.go?.main?.App;
    if (!app || !canStart.value) {
      return;
    }

    phase.value = 'installing';
    message.value = 'Starting installer';

    try {
      await app.BeginInstall(context.value.installerPath, context.value.executablePath);
      phase.value = 'handoff';
      message.value = 'Handing off';
    } catch (err) {
      phase.value = 'error';
      error.value = err instanceof Error ? err.message : 'Failed to start the installer';
    }
  };

  onMounted(() => {
    window.runtime?.EventsOn?.('bootstrapper:status', (payload: StatusPayload) => {
      if (payload.stage === 'handoff') {
        phase.value = 'handoff';
      } else if (payload.stage === 'preparing') {
        phase.value = 'installing';
      } else if (payload.stage === 'error') {
        phase.value = 'error';
        error.value = payload.message || 'Bootstrapper failed to continue the update';
      }

      if (payload.message) {
        message.value = payload.message;
      }
    });

    void loadContext();
  });

  return {
    canStart,
    context,
    error,
    message,
    phase,
    primaryActionLabel,
    startInstall,
  };
}
