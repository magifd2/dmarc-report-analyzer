import React, { useEffect, useState } from 'react';
import { useTranslation } from 'react-i18next';

interface Report {
  ID: number;
  OrgName: string;
  ReportID: string;
  DateRangeBegin: number;
  DateRangeEnd: number;
  Domain: string;
  P: string;
  SP: string;
  PCT: number;
}

const ReportList: React.FC = () => {
  const { t } = useTranslation();
  const [reports, setReports] = useState<Report[]>([]);
  const [loading, setLoading] = useState<boolean>(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    const fetchReports = async () => {
      try {
        const response = await fetch('/api/reports');
        if (!response.ok) {
          throw new Error(`HTTP error! status: ${response.status}`);
        }
        const data = await response.json();
        setReports(data.reports);
      } catch (err) {
        setError(err instanceof Error ? err.message : String(err));
      } finally {
        setLoading(false);
      }
    };

    fetchReports();
  }, []);

  if (loading) {
    return <div className="text-center text-gray-600">{t('common.loading_reports')}</div>;
  }

  if (error) {
    return <div className="text-center text-red-600">{t('common.error_prefix')}{error}</div>;
  }

  return (
    <div className="mt-8">
      <h2 className="text-xl font-bold text-gray-800 mb-4">{t('app.dmarc_reports_section_title')}</h2>
      {reports.length === 0 ? (
        <p className="text-gray-600">{t('app.no_reports_found')}</p>
      ) : (
        <div className="overflow-x-auto">
          <table className="min-w-full bg-white border border-gray-200 rounded-lg shadow-sm">
            <thead>
              <tr>
                <th className="py-2 px-4 border-b text-left text-sm font-semibold text-gray-600">{t('report_list.id_header')}</th>
                <th className="py-2 px-4 border-b text-left text-sm font-semibold text-gray-600">{t('report_list.org_name_header')}</th>
                <th className="py-2 px-4 border-b text-left text-sm font-semibold text-gray-600">{t('report_list.domain_header')}</th>
                <th className="py-2 px-4 border-b text-left text-sm font-semibold text-gray-600">{t('report_list.report_id_header')}</th>
                <th className="py-2 px-4 border-b text-left text-sm font-semibold text-gray-600">{t('report_list.date_range_header')}</th>
                <th className="py-2 px-4 border-b text-left text-sm font-semibold text-gray-600">{t('report_list.policy_header')}</th>
              </tr>
            </thead>
            <tbody>
              {reports.map((report) => (
                <tr key={report.ID} className="hover:bg-gray-50">
                  <td className="py-2 px-4 border-b text-sm text-gray-800">{report.ID}</td>
                  <td className="py-2 px-4 border-b text-sm text-gray-800">{report.OrgName}</td>
                  <td className="py-2 px-4 border-b text-sm text-gray-800">{report.Domain}</td>
                  <td className="py-2 px-4 border-b text-sm text-gray-800">{report.ReportID}</td>
                  <td className="py-2 px-4 border-b text-sm text-gray-800">
                    {new Date(report.DateRangeBegin * 1000).toLocaleDateString()} - 
                    {new Date(report.DateRangeEnd * 1000).toLocaleDateString()}
                  </td>
                  <td className="py-2 px-4 border-b text-sm text-gray-800">{report.P} ({report.PCT}%)</td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}
    </div>
  );
};

export default ReportList;
