export namespace main {
	
	export class LocalProxy {
	    protocol: string;
	    listen_ip: string;
	    listen_port: number;
	
	    static createFrom(source: any = {}) {
	        return new LocalProxy(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.protocol = source["protocol"];
	        this.listen_ip = source["listen_ip"];
	        this.listen_port = source["listen_port"];
	    }
	}
	export class UpstreamProxy {
	    protocol: string;
	    address: string;
	    username: string;
	    password: string;
	    auth_method?: string;
	
	    static createFrom(source: any = {}) {
	        return new UpstreamProxy(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.protocol = source["protocol"];
	        this.address = source["address"];
	        this.username = source["username"];
	        this.password = source["password"];
	        this.auth_method = source["auth_method"];
	    }
	}
	export class ProxyConfig {
	    id: string;
	    name: string;
	    upstream: UpstreamProxy;
	    local: LocalProxy;
	    enabled: boolean;
	    description?: string;
	
	    static createFrom(source: any = {}) {
	        return new ProxyConfig(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.upstream = this.convertValues(source["upstream"], UpstreamProxy);
	        this.local = this.convertValues(source["local"], LocalProxy);
	        this.enabled = source["enabled"];
	        this.description = source["description"];
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
	export class ProxyStatus {
	    id: string;
	    running: boolean;
	    error?: string;
	
	    static createFrom(source: any = {}) {
	        return new ProxyStatus(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.running = source["running"];
	        this.error = source["error"];
	    }
	}
	export class ProxyWithStatus {
	    id: string;
	    name: string;
	    upstream: UpstreamProxy;
	    local: LocalProxy;
	    enabled: boolean;
	    description?: string;
	    running: boolean;
	
	    static createFrom(source: any = {}) {
	        return new ProxyWithStatus(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.upstream = this.convertValues(source["upstream"], UpstreamProxy);
	        this.local = this.convertValues(source["local"], LocalProxy);
	        this.enabled = source["enabled"];
	        this.description = source["description"];
	        this.running = source["running"];
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

