/**
 * PapaParse helpers for client-side CSV preview and parsing.
 * This ensures consistent, RFC-compliant CSV handling.
 */
import Papa from 'papaparse';

/**
 * Parses a CSV string to an array of objects.
 * @param csvString The CSV content to parse.
 * @returns A promise that resolves with the parsed data.
 */
export function parseCsv(csvString: string): Promise<Papa.ParseResult<any>> {
  return new Promise((resolve, reject) => {
    Papa.parse(csvString, {
      header: true,
      skipEmptyLines: true,
      complete: (results) => resolve(results),
      error: (error: any) => reject(error),
    });
  });
}

/**
 * Converts an array of objects to a CSV string.
 * @param data The array of objects to unparse.
 * @returns The generated CSV string.
 */
export function unparseCsv(data: any[]): string {
  return Papa.unparse(data);
}
