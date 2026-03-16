export namespace main {
	
	export class LaunchContext {
	    mode: string;
	    version: string;
	    installerPath: string;
	    installerUrl: string;
	    executablePath: string;
	    installDir: string;
	
	    static createFrom(source: any = {}) {
	        return new LaunchContext(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.mode = source["mode"];
	        this.version = source["version"];
	        this.installerPath = source["installerPath"];
	        this.installerUrl = source["installerUrl"];
	        this.executablePath = source["executablePath"];
	        this.installDir = source["installDir"];
	    }
	}

}

