import React from 'react';
// import Settings from '../../assets/icons/settings.svg';
import Docs from '../../../assets/icons/docs.svg';
import RightArrow from '../../../assets/icons/rightArrow.svg';
import Github from '../../../assets/icons/github.svg';
import Twitter from '../../../assets/icons/twitter.svg';
import Slack from '../../../assets/icons/slack.svg';
import Email from '../../../assets/icons/email.svg';
import './LeftNavBar.scss';

const LeftNavBar = (): JSX.Element => {
  return (
    <section className="left-nav-bar">
      <div className="top-div">
        <h1>Polaris</h1>
      </div>
      <div className="bottom-div">
        <h2 className="links-title">Application</h2>
        {/* TODO: no settings in Polaris? */}
        {/* <div className="nav-link-section">
          <img src={Settings} alt='gear icon' />
          <h3 className="link-name">Settings</h3>
        </div> */}
        <a className="nav-link-section" href="https://polaris.docs.fairwinds.com/" target="_blank" rel="noreferrer">
          <img src={Docs} alt='doc icon' />
          <h3 className="link-name docs">Docs</h3>
          <img src={RightArrow} alt='arrow pointing right' />
        </a>
        <div className="external-links">
          <a href="https://github.com/FairwindsOps/polaris" target="_blank" rel="noreferrer"><img src={Github} alt='github logo' /></a>
          <a href="https://twitter.com/fairwindsops" target="_blank" rel="noreferrer"><img src={Twitter} alt='twitter logo' /></a>
          <a href="https://join.slack.com/t/fairwindscommunity/shared_invite/zt-e3c6vj4l-3lIH6dvKqzWII5fSSFDi1g" target="_blank" rel="noreferrer"><img src={Slack} alt='slack logo' /></a>
          <a href="https://www.fairwinds.com/fairwinds-newsletter" target="_blank" rel="noreferrer"><img src={Email} alt='envelope for newsletter' /></a>
        </div>
        <a href="https://github.com/fairwindsops/polaris/issues" target="_blank" rel="noreferrer" className="feedback">Feedback</a>
      </div>
    </section>
  )
}

export default LeftNavBar;