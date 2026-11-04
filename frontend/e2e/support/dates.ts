// Calendar dates relative to today, in the local time zone of the machine running the tests.
function daysFromToday(days: number): Date {
  const date = new Date();
  date.setDate(date.getDate() + days);
  return date;
}

// YYYY-MM-DD, the value format of <input type="date">.
export function inputDate(days: number): string {
  const date = daysFromToday(days);
  return [
    date.getFullYear(),
    String(date.getMonth() + 1).padStart(2, "0"),
    String(date.getDate()).padStart(2, "0"),
  ].join("-");
}

// dd Mon yyyy, the format the app shows dates in.
export function displayDate(days: number): string {
  return daysFromToday(days).toLocaleDateString("en-GB", {
    day: "2-digit",
    month: "short",
    year: "numeric",
  });
}
