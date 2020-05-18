package search_test

//import (
//    . "github.com/onsi/ginkgo"
//    . "github.com/onsi/gomega"
//    "github.com/pantheon-systems/search-secrets/pkg/dev"
//    . "github.com/pantheon-systems/search-secrets/pkg/search"
//    "github.com/pantheon-systems/search-secrets/pkg/search/contract"
//    "testing"
//)
//
//var _ = Describe("Search Secrets Functional tests", func() {
//
//    Describe("Parameters", func() {
//
//        Context("If no arguments are passed", func() {
//
//            It("we should get an error", func() {
//                //writer := resultWriter{}
//                //job := buildJob()
//                //
//                //// Work
//                //job.Perform(writer)
//                //
//                //Expect(writer.results).To(HaveLen(1))
//            })
//        })
//    })
//})
//
//type resultWriter struct {
//    results []*contract.JobResult
//}
//
//func (r *resultWriter) WriteResult(result *contract.JobResult) (err error) {
//    r.results = append(r.results, result)
//    return
//}
//
//func buildJob() (result Worker) {
//    //var err error
//    //
//    //logger := logrus.New()
//    //logger.Out = ioutil.Discard
//    //log := logrus.NewEntry(logger)
//    //
//    //// Code filter res
//    //// Code filter
//    //var codeWhitelist *CodeWhitelist
//    //codeWhitelist = NewCodeWhitelist(manip.NewRegexpSet(nil))
//    //
//    //// Git
//    //var git *gitpkg.Git
//    //git = gitpkg.New(log)
//    //
//    //cloneDir := "/Users/mattalexander/go/src/github.com/pantheon-systems/search-secrets/" +
//    //    "_pantheon/output2/source/titan-mt"
//    //var repository *gitpkg.Repository
//    //repository, err = git.OpenRepository(cloneDir)
//    //Expect(err).To(Not(HaveOccurred()))
//    //
//    //var commitHashes []string
//    //commitHashes = []string{"31c683f2f582f0027d5794b2a173adefacad326e"}
//    //
//    //var oldest string
//    //oldest = ""
//    //
//    //// SearchTarget
//    //searchTarget := NewJob(repository, commitHashes, oldest)
//    //
//    //// ProcessorI
//    //processor := setter.NewProcessor("test-setter-proc", codeWhitelist)
//    //
//    //var processors []ProcessorI
//    //processors = append(processors, processor)
//    //
//    //var pathFilter *manip.RegexpFilter
//    //pathFilter = manip.NewRegexpFilterFromStringSlicesMustCompile(nil, nil)
//    //
//    //var fileChangeFilter *gitpkg.FileChangeFilter
//    //fileChangeFilter = gitpkg.NewFileChangeFilter(pathFilter, true, true, true)
//    //
//    //var secretIDFilter manip.Filter
//    //secretIDFilter = manip.NewFilterFromStringSlices(nil, nil)
//    //
//    //name := "test-search-worker"
//    //
//    //// func NewWorker(name string, searchTarget *Job, processors []ProcessorI, fileChangeFilter *gitpkg.FileChangeFilter,
//    ////    secretIDFilter manip.Filter, bar *progress.Bar, log logg.Logg) SearchWorker {
//    //var bar *progress.Bar
//    //bar = nil
//    //result = NewWorker(name, searchTarget, processors, fileChangeFilter, secretIDFilter, bar, log)
//
//    return
//}
//
//func TestErrors(t *testing.T) {
//    dev.RunningTests = true
//    RegisterFailHandler(Fail)
//    RunSpecs(t, "Search Job Suite")
//}
