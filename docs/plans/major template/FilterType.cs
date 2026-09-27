namespace CHECKIN.Common
{
    public class StringFilter
    {
        public string Equal { get; set; }
        public string NotEqual { get; set; }
        /// <summary>
        /// x.contains(string)
        /// </summary>
        public string Contain { get; set; }
        /// <summary>
        /// !x.contains(string)
        /// </summary>
        public string NotContain { get; set; }
        /// <summary>
        /// string.contains(x)
        /// </summary>
        public string ReverseContain { get; set; }
        /// <summary>
        /// !string.Contains(x)
        /// </summary>
        public string ReverseNotContain { get; set; }
        /// <summary>
        /// string.Contains(x) || x.Contains(string)
        /// </summary>
        public string CombineContain { get; set; }
        /// <summary>
        /// x.Startswith(string)
        /// </summary>
        public string StartWith { get; set; }
        /// <summary>
        /// !x.Startswith(string)
        /// </summary>
        public string NotStartWith { get; set; }
        /// <summary>
        /// string.Startswith(x)
        /// </summary>
        public string ReverseStartWith { get; set; }
        /// <summary>
        /// string.Startswith(x)
        /// </summary>
        public string ReverseNotStartWith { get; set; }
        /// <summary>
        /// x.Startswith(string) || string.Startswith(x)
        /// </summary>
        public string CombineStartWith { get; set; }
        /// <summary>
        /// x.Endswith(string)
        /// </summary>
        public string EndWith { get; set; }
        /// <summary>
        /// !x.Endswith(string)
        /// </summary>
        public string NotEndWith { get; set; }
        /// <summary>
        /// string.EndsWith(x)
        /// </summary>
        public string ReverseEndWith { get; set; }
        /// <summary>
        /// string.EndsWith(x)
        /// </summary>
        public string ReverseNotEndWith { get; set; }
        /// <summary>
        /// x.EndsWith(string) || string.EndsWith(x)
        /// </summary>
        public string CombineEndWith { get; set; }

        public List<string> In { get; set; }
        public List<string> NotIn { get; set; }
        public StringFilter ToLower()
        {
            Equal = Equal?.ToLower();
            NotEqual = NotEqual?.ToLower();
            Contain = Contain?.ToLower();
            NotContain = NotContain?.ToLower();
            ReverseContain = ReverseContain?.ToLower();
            ReverseNotContain = ReverseNotContain?.ToLower();
            CombineContain = CombineContain?.ToLower();
            StartWith = StartWith?.ToLower();
            NotStartWith = NotStartWith?.ToLower();
            ReverseStartWith = ReverseStartWith?.ToLower();
            ReverseNotStartWith = ReverseNotStartWith?.ToLower();
            CombineStartWith = CombineStartWith?.ToLower();
            EndWith = EndWith?.ToLower();
            NotEndWith = NotEndWith?.ToLower();
            ReverseEndWith = ReverseEndWith?.ToLower();
            ReverseNotEndWith = ReverseNotEndWith?.ToLower();
            CombineEndWith = CombineEndWith?.ToLower();
            In = In?.Select(x => x.ToLower()).ToList();
            NotIn = NotIn?.Select(x => x.ToLower()).ToList();
            return this;
        }

        public StringFilter ToUpper()
        {
            Equal = Equal?.ToUpper();
            NotEqual = NotEqual?.ToUpper();
            Contain = Contain?.ToUpper();
            NotContain = NotContain?.ToUpper();
            ReverseContain = ReverseContain?.ToUpper();
            ReverseNotContain = ReverseNotContain?.ToUpper();
            CombineContain = CombineContain?.ToUpper();
            StartWith = StartWith?.ToUpper();
            NotStartWith = NotStartWith?.ToUpper();
            ReverseStartWith = ReverseStartWith?.ToUpper();
            ReverseNotStartWith = ReverseNotStartWith?.ToUpper();
            CombineStartWith = CombineStartWith?.ToUpper();
            EndWith = EndWith?.ToUpper();
            NotEndWith = NotEndWith?.ToUpper();
            ReverseEndWith = ReverseEndWith?.ToUpper();
            ReverseNotEndWith = ReverseNotEndWith?.ToUpper();
            CombineEndWith = CombineEndWith?.ToUpper();
            In = In?.Select(x => x.ToUpper()).ToList();
            NotIn = NotIn?.Select(x => x.ToUpper()).ToList();
            return this;
        }
        public void TrimString()
        {
            Equal = Equal?.Trim();
            NotEqual = NotEqual?.Trim();
            Contain = Contain?.Trim();
            NotContain = NotContain?.Trim();
            ReverseContain = ReverseContain?.Trim();
            ReverseNotContain = ReverseNotContain?.Trim();
            CombineContain = CombineContain?.Trim();
            StartWith = StartWith?.Trim();
            NotStartWith = NotStartWith?.Trim();
            ReverseStartWith = ReverseStartWith?.Trim();
            ReverseNotStartWith = ReverseNotStartWith?.Trim();
            CombineStartWith = CombineStartWith?.Trim();
            EndWith = EndWith?.Trim();
            NotEndWith = NotEndWith?.Trim();
            ReverseEndWith = ReverseEndWith?.Trim();
            ReverseNotEndWith = ReverseNotEndWith?.Trim();
            CombineEndWith = CombineEndWith?.Trim();
            In = In?.Select(x => x.Trim()).ToList();
            NotIn = NotIn?.Select(x => x.Trim()).ToList();
        }
        public bool HasValue
        {
            get
            {
                return !string.IsNullOrWhiteSpace(Equal) ||
                    !string.IsNullOrWhiteSpace(NotEqual) ||
                    !string.IsNullOrWhiteSpace(Contain) ||
                    !string.IsNullOrWhiteSpace(NotContain) ||
                    !string.IsNullOrWhiteSpace(ReverseContain) ||
                    !string.IsNullOrWhiteSpace(ReverseNotContain) ||
                    !string.IsNullOrWhiteSpace(CombineContain) ||
                    !string.IsNullOrWhiteSpace(StartWith) ||
                    !string.IsNullOrWhiteSpace(NotStartWith) ||
                    !string.IsNullOrWhiteSpace(ReverseStartWith) ||
                    !string.IsNullOrWhiteSpace(ReverseNotStartWith) ||
                    !string.IsNullOrWhiteSpace(CombineStartWith) ||
                    !string.IsNullOrWhiteSpace(EndWith) ||
                    !string.IsNullOrWhiteSpace(NotEndWith) ||
                    !string.IsNullOrWhiteSpace(ReverseEndWith) ||
                    !string.IsNullOrWhiteSpace(ReverseNotEndWith) ||
                    !string.IsNullOrWhiteSpace(CombineEndWith) ||
                    (In != null && In.Count > 0) ||
                    (NotIn != null && NotIn.Count > 0);
            }
        }
    }

    public class LongFilter
    {
        public long? Equal { get; set; }
        public long? NotEqual { get; set; }
        public long? Less { get; set; }
        public long? LessEqual { get; set; }
        public long? Greater { get; set; }
        public long? GreaterEqual { get; set; }
        public List<long> In { get; set; }
        public List<long> NotIn { get; set; }
        public bool HasValue
        {
            get
            {
                return Equal.HasValue ||
                    NotEqual.HasValue ||
                    Less.HasValue ||
                    LessEqual.HasValue ||
                    Greater.HasValue ||
                    GreaterEqual.HasValue ||
                    (In != null && In.Count > 0) ||
                    (NotIn != null && NotIn.Count > 0);
            }
        }
    }

    public class DecimalFilter
    {
        public decimal? Equal { get; set; }
        public decimal? NotEqual { get; set; }
        public decimal? Less { get; set; }
        public decimal? LessEqual { get; set; }
        public decimal? Greater { get; set; }
        public decimal? GreaterEqual { get; set; }
        public List<decimal> In { get; set; }
        public List<decimal> NotIn { get; set; }
        public bool HasValue
        {
            get
            {
                return Equal.HasValue ||
                    NotEqual.HasValue ||
                    Less.HasValue ||
                    LessEqual.HasValue ||
                    Greater.HasValue ||
                    GreaterEqual.HasValue ||
                    (In != null && In.Count > 0) ||
                    (NotIn != null && NotIn.Count > 0);
            }
        }
    }

    public class DateFilter
    {
        public DateTime? Equal { get; set; }
        public DateTime? NotEqual { get; set; }
        public DateTime? Less { get; set; }
        public DateTime? LessEqual { get; set; }
        public DateTime? Greater { get; set; }
        public DateTime? GreaterEqual { get; set; }
        public bool HasValue
        {
            get
            {
                return Equal.HasValue ||
                    NotEqual.HasValue ||
                    Less.HasValue ||
                    LessEqual.HasValue ||
                    Greater.HasValue ||
                    GreaterEqual.HasValue;
            }
        }
    }

    public class GuidFilter
    {
        public Guid? Equal { get; set; }
        public Guid? NotEqual { get; set; }
        public List<Guid> In { get; set; }
        public List<Guid> NotIn { get; set; }

        public bool HasValue
        {
            get
            {
                return Equal.HasValue ||
                    NotEqual.HasValue ||
                    (In != null && In.Count > 0) ||
                    (NotIn != null && NotIn.Count > 0);
            }
        }
    }

    public class IdFilter
    {
        public long? Equal { get; set; }
        public long? NotEqual { get; set; }
        public List<long> In { get; set; }
        public List<long> NotIn { get; set; }

        public bool HasValue
        {
            get
            {
                return Equal.HasValue ||
                    NotEqual.HasValue ||
                    (In != null && In.Count > 0) ||
                    (NotIn != null && NotIn.Count > 0);
            }
        }
    }
}

