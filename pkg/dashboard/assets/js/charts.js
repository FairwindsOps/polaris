$(function () {
  var clusterChart = new Chart("clusterScoreChart", {
    type: 'doughnut',
    data: {
      labels: ["Passing", "Warning", "Error"],
      datasets: [{
        data: [
          fairwindsAuditData.ClusterSummary.Results.Totals.Successes,
          fairwindsAuditData.ClusterSummary.Results.Totals.Warnings,
          fairwindsAuditData.ClusterSummary.Results.Totals.Errors,
        ],
        backgroundColor: ['#8BD2DC', '#f26c21', '#a11f4c'],
      }]
    },
    options: {
      responsive: false,
      legend: {
        display: false,
      },
    }
  });
});
