import React, { useState, useEffect } from 'react';
import LeftNavBar from './components/Navigation/LeftBar/LeftNavBar';
import TopNavBar from './components/Navigation/TopBar/TopNavBar';

import './App.scss';
import data from './data.json';

function App() {
  const [pageDisplay, setPageDisplay] = useState<string>('dashboard');
  const [namespaces, setNamespaces] = useState<string[]>([]);
  const [selectedNamespace, setSelectedNamespace] = useState<string>('');

  useEffect(() => {
    const allNamespaces: Set<string> = new Set();
    data.Results.forEach((result) => {
      allNamespaces.add(result.Namespace);
    });
    setNamespaces(Array.from(allNamespaces));
    setSelectedNamespace('');
  }, []);

  return (
    <div className="App">
      <LeftNavBar />
      <div id="mainContainer">
        <TopNavBar
          pageDisplay={pageDisplay}
          setPageDisplay={setPageDisplay}
          namespaces={namespaces}
          setSelected={setSelectedNamespace}
        />
        <main className="app-content">
          {pageDisplay === 'dashboard' && <h1>Polaris Dashboard</h1>}
          {pageDisplay === 'namespaces' && (
            <h1>{selectedNamespace} Namespace</h1>
          )}
        </main>
      </div>
    </div>
  );
}

export default App;
