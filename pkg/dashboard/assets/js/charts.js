$(function () {
  var clusterChart = new Chart("clusterScoreChart", {
    type: 'doughnut',
    data: {
      labels: ["Passing", "Warning", "Error"],
      datasets: [{
        data: [
          polarisAuditData.ClusterSummary.Results.Totals.Successes,
          polarisAuditData.ClusterSummary.Results.Totals.Warnings,
          polarisAuditData.ClusterSummary.Results.Totals.Errors,
        ],
        backgroundColor: ['#8BD2DC', '#f26c21', '#a11f4c'],
      }]
    },
    options: {
      responsive: false,
      cutoutPercentage: 60,
      legend: {
        display: false,
      },
    }
  });
});
