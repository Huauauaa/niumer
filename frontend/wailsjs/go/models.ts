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

}

