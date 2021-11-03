import React, { useState } from 'react';
import { Nav, NavDropdown } from 'react-bootstrap';
import './TopNavBar.scss';

type NavProps = {
  pageDisplay: string,
  setPageDisplay: React.Dispatch<React.SetStateAction<string>>,
  namespaces: string[],
  setSelected: React.Dispatch<React.SetStateAction<string>>,
}

const TopNavBar = ({ pageDisplay, setPageDisplay, namespaces, setSelected }: NavProps): JSX.Element => {
  const [selectedNamespace, setSelectedNamespace] = useState<string>('All Namespaces');

  const handleNamespaceSelection = (namespace: React.ReactEventHandler<HTMLDivElement>) => {
    if (namespace.toString() === 'All Namespaces') {
      setSelectedNamespace('All Namespace');
    } else if (namespace) {
      setSelectedNamespace(namespace.toString());
    }
    setPageDisplay('namespaces')
    setSelected(namespace.toString());
  };

  const NavItems = (): JSX.Element => {
    return (
      <>
        {namespaces.map((item) => {
          return (
            <NavDropdown.Item key={item} eventKey={item}>
              {item}
            </NavDropdown.Item>
          );
        })}
      </>
    );
  };

  return (
    <div className="top-nav-bar">
      <Nav variant="tabs" className="top-nav">
        <NavDropdown title={selectedNamespace} id="nav-dropdown" active={pageDisplay === 'namespaces'} onSelect={(namespace: any) => handleNamespaceSelection(namespace)}>
          <NavDropdown.Item key={'all-namespaces'} eventKey={'all-namespaces'}>
            All Namespaces
          </NavDropdown.Item>
          <NavItems />
        </NavDropdown>
        <Nav.Item onClick={() => setPageDisplay('dashboard')}>
          <Nav.Link active={pageDisplay === 'dashboard'}>
            Dashboard
          </Nav.Link>
        </Nav.Item>
      </Nav>
    </div>
  )
}

export default TopNavBar;