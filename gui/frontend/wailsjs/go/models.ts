export namespace history {
	
	export class Entry {
	    src: string;
	    dest: string;
	    action: string;
	
	    static createFrom(source: any = {}) {
	        return new Entry(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.src = source["src"];
	        this.dest = source["dest"];
	        this.action = source["action"];
	    }
	}
	export class Run {
	    id: number;
	    // Go type: time
	    timestamp: any;
	    folders: string[];
	    entries: Entry[];
	
	    static createFrom(source: any = {}) {
	        return new Run(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.timestamp = this.convertValues(source["timestamp"], null);
	        this.folders = source["folders"];
	        this.entries = this.convertValues(source["entries"], Entry);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}

}

export namespace main {
	
	export class runResult {
	    source: string;
	    destination: string;
	    skipped: boolean;
	    skip_reason: string;
	    action: string;
	
	    static createFrom(source: any = {}) {
	        return new runResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.source = source["source"];
	        this.destination = source["destination"];
	        this.skipped = source["skipped"];
	        this.skip_reason = source["skip_reason"];
	        this.action = source["action"];
	    }
	}
	export class scheduleView {
	    name: string;
	    status: string;
	    service: boolean;
	    cron: string;
	    folders: string[];
	    config: string;
	    // Go type: time
	    last_run: any;
	    last_files: number;
	
	    static createFrom(source: any = {}) {
	        return new scheduleView(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.status = source["status"];
	        this.service = source["service"];
	        this.cron = source["cron"];
	        this.folders = source["folders"];
	        this.config = source["config"];
	        this.last_run = this.convertValues(source["last_run"], null);
	        this.last_files = source["last_files"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}

}

