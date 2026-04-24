export namespace main {
	
	export class AttendanceRecord {
	    id: number;
	    creationDate: string;
	    createdBy: string;
	    lastUpdateDate: string;
	    lastUpdatedBy: string;
	    originalId: string;
	    hrId: number;
	    dataSource: string;
	    clockInReason: string;
	    attendanceDate: string;
	    clockInDate: string;
	    clockInTime: string;
	    dayId: string;
	    clockingInSequenceNumber: number;
	    earlyClockInTime: string;
	    lateClockInTime: string;
	    clockInType: string;
	    earlyClockInType: string;
	    lateClockInType: string;
	    attendanceStatus: string;
	    minuteNumber: string;
	    hourNumber: string;
	    attendProcessId: string;
	    workDay: string;
	    attendanceStatusCode: string;
	    earlyClockInReason: string;
	    lateClockInReason: string;
	    earlyClockTag: string;
	    lateClockTag: string;
	    effectiveWorkHours: number;
	
	    static createFrom(source: any = {}) {
	        return new AttendanceRecord(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.creationDate = source["creationDate"];
	        this.createdBy = source["createdBy"];
	        this.lastUpdateDate = source["lastUpdateDate"];
	        this.lastUpdatedBy = source["lastUpdatedBy"];
	        this.originalId = source["originalId"];
	        this.hrId = source["hrId"];
	        this.dataSource = source["dataSource"];
	        this.clockInReason = source["clockInReason"];
	        this.attendanceDate = source["attendanceDate"];
	        this.clockInDate = source["clockInDate"];
	        this.clockInTime = source["clockInTime"];
	        this.dayId = source["dayId"];
	        this.clockingInSequenceNumber = source["clockingInSequenceNumber"];
	        this.earlyClockInTime = source["earlyClockInTime"];
	        this.lateClockInTime = source["lateClockInTime"];
	        this.clockInType = source["clockInType"];
	        this.earlyClockInType = source["earlyClockInType"];
	        this.lateClockInType = source["lateClockInType"];
	        this.attendanceStatus = source["attendanceStatus"];
	        this.minuteNumber = source["minuteNumber"];
	        this.hourNumber = source["hourNumber"];
	        this.attendProcessId = source["attendProcessId"];
	        this.workDay = source["workDay"];
	        this.attendanceStatusCode = source["attendanceStatusCode"];
	        this.earlyClockInReason = source["earlyClockInReason"];
	        this.lateClockInReason = source["lateClockInReason"];
	        this.earlyClockTag = source["earlyClockTag"];
	        this.lateClockTag = source["lateClockTag"];
	        this.effectiveWorkHours = source["effectiveWorkHours"];
	    }
	}
	export class CustomReminder {
	    id: string;
	    name: string;
	    date: string;
	
	    static createFrom(source: any = {}) {
	        return new CustomReminder(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.date = source["date"];
	    }
	}
	export class PullRequestListItem {
	    id: number;
	    number: number;
	    url: string;
	    title: string;
	    author: string;
	    sourceBranch: string;
	    targetBranch: string;
	    state: string;
	    createdAt: string;
	    updatedAt: string;
	
	    static createFrom(source: any = {}) {
	        return new PullRequestListItem(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.number = source["number"];
	        this.url = source["url"];
	        this.title = source["title"];
	        this.author = source["author"];
	        this.sourceBranch = source["sourceBranch"];
	        this.targetBranch = source["targetBranch"];
	        this.state = source["state"];
	        this.createdAt = source["createdAt"];
	        this.updatedAt = source["updatedAt"];
	    }
	}
	export class PullRequestListResponse {
	    items: PullRequestListItem[];
	    total: number;
	    page: number;
	    pageSize: number;
	    totalPages: number;
	
	    static createFrom(source: any = {}) {
	        return new PullRequestListResponse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.items = this.convertValues(source["items"], PullRequestListItem);
	        this.total = source["total"];
	        this.page = source["page"];
	        this.pageSize = source["pageSize"];
	        this.totalPages = source["totalPages"];
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
	export class WorkHourUserProfileView {
	    userAccount: string;
	    hrId: number;
	    shiftNameZh: string;
	    updatedAt: string;
	    userInfoJson: string;
	
	    static createFrom(source: any = {}) {
	        return new WorkHourUserProfileView(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.userAccount = source["userAccount"];
	        this.hrId = source["hrId"];
	        this.shiftNameZh = source["shiftNameZh"];
	        this.updatedAt = source["updatedAt"];
	        this.userInfoJson = source["userInfoJson"];
	    }
	}

}

