/// <reference types="vite/client" />

interface BootstrapperLaunchContext {
  mode: string;
  version: string;
  installerPath: string;
  executablePath: string;
  installDir: string;
}

interface BootstrapperAppApi {
  GetLaunchContext: () => Promise<BootstrapperLaunchContext>;
  BeginInstall: (installerPath?: string, executablePath?: string) => Promise<void>;
}

interface RuntimeApi {
  EventsOn?: (eventName: string, callback: (payload: any) => void) => void;
}

interface Window {
  go?: {
    main?: {
      App?: BootstrapperAppApi;
    };
  };
  runtime?: RuntimeApi;
}
