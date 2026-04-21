/** Matches main.AttendanceRecord JSON from Go (attendance_records table). */
export type AttendanceRecord = {
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
  /** 有效工时（小时）：两段打卡均存在时，与 8–12、13:30–17:30、18–24 时段的交集；否则 0 */
  effectiveWorkHours: number;
};
