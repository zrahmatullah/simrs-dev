import * as XLSX from 'xlsx'

/**
 * Utility to export tabular data directly to genuine Microsoft Excel (.xlsx) file
 * with auto column width adjustment and clean formatting.
 */
export function exportToExcel(filename: string, sheetName: string, data: any[]) {
  if (!data || data.length === 0) {
    alert('Tidak ada data yang tersedia untuk diunduh ke Excel.')
    return
  }

  // 1. Create a new workbook
  const wb = XLSX.utils.book_new()

  // 2. Convert json data to worksheet
  const ws = XLSX.utils.json_to_sheet(data)

  // 3. Auto-fit column widths based on cell content
  const colKeys = Object.keys(data[0])
  const colWidths = colKeys.map(key => {
    const maxLen = Math.max(
      key.length,
      ...data.map(row => {
        const val = row[key]
        if (val === null || val === undefined) return 0
        return String(val).length
      })
    )
    return { wch: Math.min(Math.max(maxLen + 3, 12), 45) }
  })
  ws['!cols'] = colWidths

  // 4. Append sheet to workbook and trigger browser download
  XLSX.utils.book_append_sheet(wb, ws, sheetName.substring(0, 31))
  const cleanFilename = filename.endsWith('.xlsx') ? filename : `${filename}.xlsx`
  XLSX.writeFile(wb, cleanFilename)
}
